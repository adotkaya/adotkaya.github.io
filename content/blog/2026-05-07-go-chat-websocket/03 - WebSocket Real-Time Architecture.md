---
title: "03 — WebSocket Real-Time Architecture"
date: 2026-05-07
tags: [go, websockets, concurrency, real-time, hub-pattern]
---

# 03 — WebSocket Real-Time Architecture

The WebSocket layer is the beating heart of `go-chat`. It transforms a simple HTTP API into a living, breathing real-time system where messages flow instantly between connected clients. This is also the most technically interesting part of the project — the part interviewers will ask about when they see "real-time chat" on a resume.

This note explores the **Hub pattern**, the goroutine model, and the concurrency decisions that make broadcasting to hundreds of clients both fast and safe.

## The Problem

HTTP is request-response. A client asks, the server answers, and the connection closes. Chat doesn't work like that. When User A sends a message, User B needs to receive it *without* asking for it. The naive approach — long polling — is wasteful: clients hammer the server with "any new messages?" requests, burning CPU and bandwidth.

WebSockets solve this by upgrading an HTTP connection into a persistent, full-duplex TCP socket. Once upgraded, either side can send data at any time. The challenge shifts from "how do we talk" to "how do we manage thousands of simultaneous conversations without leaking memory or blocking goroutines."

## The Hub Pattern

The `Hub` is a single goroutine that owns all WebSocket state. It is the central coordinator, the traffic cop, the post office. Every client connection flows through it.

```go
type Hub struct {
    rooms      map[int64]map[*Client]bool  // roomID -> set of clients
    broadcast  chan *Message                // inbound messages from clients
    register   chan *Client                 // new connections
    unregister chan *Client                 // disconnections
    store      store.Storage                // persistence layer
}
```

Why a single goroutine? Because it eliminates the need for mutexes. The Hub owns the `rooms` map, and nothing else touches it. All access happens through channels, which are Go's idiomatic way to share memory by communicating. This is the canonical pattern from the Gorilla WebSocket `chat` example, refined for production.

### The Event Loop

```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.registerClient(client)
        case client := <-h.unregister:
            h.unregisterClient(client)
        case message := <-h.broadcast:
            h.handleBroadcast(message)
        }
    }
}
```

This `for { select {} }` loop is the entire Hub. It waits on three channels:

1. **`register`**: A new client connected. Add them to their room's client set.
2. **`unregister`**: A client disconnected. Remove them, close their send channel, clean up empty rooms.
3. **`broadcast`**: A message arrived. Persist it, then fan it out to every client in the room.

The `select` statement guarantees that handling one event never blocks another. A slow client cannot prevent a new client from registering.

## The Client: Two Goroutines Per Connection

Every WebSocket connection spawns exactly two goroutines:

```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte        // buffered outbound messages
    userID   int64
    username string
    roomID   int64
}
```

### readPump: Inbound Messages

`readPump` runs in its own goroutine and continuously reads from the WebSocket:

```go
func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, ...) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        msg := &Message{
            RoomID:   c.roomID,
            UserID:   c.userID,
            Username: c.username,
            Content:  string(message),
            Type:     "message",
        }
        c.hub.broadcast <- msg
    }
}
```

The `defer` is critical. When `readPump` exits — whether from an error, a close frame, or a timeout — it sends the client to `h.unregister` and closes the connection. This guarantees cleanup even if the client vanishes without saying goodbye (power loss, network drop, browser crash).

The **ping/pong mechanism** detects dead connections. The server sends pings periodically. If the client doesn't respond with a pong within `pongWait` (60 seconds), `ReadMessage` returns a timeout error and `readPump` exits. Without this, a client that disappears silently would leak a goroutine and a slot in the `rooms` map indefinitely.

### writePump: Outbound Messages

`writePump` runs in a second goroutine and handles sending messages to the client:

```go
func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, [])
                return
            }

            w, _ := c.conn.NextWriter(websocket.TextMessage)
            w.Write(message)

            // Batch queued messages into a single frame
            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.send)
            }
            w.Close()

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            c.conn.WriteMessage(websocket.PingMessage, nil)
        }
    }
}
```

Why two goroutines? Because the WebSocket library is **not thread-safe for concurrent reads and writes**. By dedicating one goroutine exclusively to reading and one exclusively to writing, we eliminate all races on the connection without a single mutex.

The **message batching** optimization is subtle but important. If multiple messages arrive while `writePump` is busy, they queue in `c.send`. On the next iteration, instead of writing them one at a time, we drain the queue and batch them into a single WebSocket frame separated by newlines. Fewer frames = fewer syscalls = lower latency under load.

## Broadcasting: The Fan-Out

When a message hits the `broadcast` channel, the Hub does two things: persist and propagate.

```go
func (h *Hub) handleBroadcast(message *Message) {
    if message.Type == "message" {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        dbMessage := &store.Message{
            RoomID:  message.RoomID,
            UserID:  message.UserID,
            Content: message.Content,
        }
        if err := h.store.Messages.Create(ctx, dbMessage); err != nil {
            log.Printf("Failed to save message: %v", err)
        }
    }

    h.broadcastToRoom(message.RoomID, message)
}
```

Database persistence happens inline, in the Hub goroutine. This is a deliberate trade-off. The Hub is single-threaded, so a slow database write would block all other events. The 5-second timeout mitigates this — if the database hangs, we log the error and continue broadcasting. In a production system, you'd likely move persistence to a background worker or use an async queue.

The actual fan-out:

```go
func (h *Hub) broadcastToRoom(roomID int64, message *Message) {
    clients := h.rooms[roomID]
    jsonMessage, _ := json.Marshal(message)

    for client := range clients {
        select {
        case client.send <- jsonMessage:
            // delivered to client's writePump
        default:
            // client's buffer is full — likely dead or slow
            close(client.send)
            delete(clients, client)
        }
    }
}
```

Three critical decisions here:

1. **Marshal once.** We call `json.Marshal` a single time, then send the same `[]byte` to every client. If 500 people are in a room, we serialize once, not 500 times.

2. **Non-blocking send.** The `select` with a `default` case means we never wait for a slow client. If `client.send` is full, we close the channel and delete the client. One bad connection cannot stall the entire room.

3. **No mutex on `rooms`.** Because the Hub is a single goroutine, the `rooms` map is only accessed from one thread. No locks, no contention, no deadlocks.

## Join and Leave Notifications

When a client registers or unregisters, the Hub synthesizes system messages:

```go
joinMessage := &Message{
    RoomID:   client.roomID,
    UserID:   client.userID,
    Username: client.username,
    Content:  client.username + " joined the room",
    Type:     "join",
}
h.broadcastToRoom(client.roomID, joinMessage)
```

These have `Type: "join"` or `Type: "leave"` instead of `Type: "message"`. The persistence layer skips them — they are ephemeral UI notifications, not chat history. The frontend uses the type field to style them differently (typically gray, italic, centered).

## Memory Model and Trade-Offs

| Decision | Choice | Trade-Off |
|----------|--------|-----------|
| Hub goroutines | 1 | Simple, no locks. Becomes bottleneck at extreme scale. |
| Client goroutines | 2 per connection | Clean separation, no races on `websocket.Conn`. Higher memory per user. |
| Send buffer | 256 messages | Prevents slow clients from blocking. Drops messages if buffer fills. |
| Broadcast marshaling | Once per message | Saves CPU. Requires immutable `[]byte`. |
| Database writes | Inline in Hub | Simple. Risk: slow DB stalls the Hub. Mitigated by timeout. |

At 1,000 concurrent users, this costs roughly ~3,000 goroutines (1 Hub + 2 per client). In Go, goroutines are cheap — a few kilobytes each — so this is well within the capabilities of a single server. The real limit is the OS file descriptor count and network bandwidth, not Go's scheduler.

## What This Is Not (Yet)

This is a single-node design. The Hub lives in one process, and clients must connect to that process. If you want to run multiple server instances behind a load balancer, this architecture breaks: User A connects to Server 1, User B connects to Server 2, and the Hubs cannot see each other.

The standard solution is a **message bus** like Redis Pub/Sub or NATS. Each server subscribes to a room channel. When Server 1 receives a message, it publishes to Redis, and all servers — including Server 2 — receive it and broadcast to their local clients. The Hub's fan-out logic stays identical; you just replace the `broadcast` channel with a bus subscription.

## Interview Hook

**Q: "How would you scale this to multiple servers?"**

A: The single-Hub design is the bottleneck. I'd introduce Redis Pub/Sub as a backplane. Each server instance maintains its own Hub for local clients, but subscribes to Redis channels per room. When a message arrives on a WebSocket, the server publishes it to Redis. All servers receive it and fan it out to their local clients. The Hub's `broadcastToRoom` logic doesn't change — the `broadcast` channel just gets fed from Redis instead of a local `readPump`.

**Q: "Why not use a mutex instead of a single Hub goroutine?"**

A: You could. But channels are Go's idiomatic concurrency primitive. A single goroutine with `select` is easier to reason about than a mutex-protected map that gets touched by thousands of goroutines. No risk of forgetting to unlock, no priority inversion, no deadlocks. The trade-off is throughput: one goroutine can only do one thing at a time. For chat, that's usually fine.

**Q: "What happens if the database write in `handleBroadcast` is slow?"**

A: The 5-second context timeout prevents indefinite blocking. If the database is down, we log the error and continue broadcasting. Messages are delivered in real-time even if persistence fails. In a production system, I'd move writes to a background queue or use a write-behind cache to decouple broadcast speed from database latency.

## Related Notes

- [[01 - Project Overview]]: Goals, tech stack, and what the system can do today.
- [[02 - Architecture and Design Patterns]]: Why interfaces and dependency injection make the Hub testable.
- [[04 - Authentication & Security]]: How JWT validation works during the WebSocket upgrade handshake.
- [[05 - Deployment, Build & Operations]]: How the Hub starts alongside the HTTP server in `main.go`.
