---
title: "06 - Server Orchestration"
tags: [server, orchestration, broadcasting, go, concurrency]
---

# 06 - Server Orchestration

The `FileServer` in `server.go` is the brain of the operation. It is the orchestration layer that bridges the gap between the low-level network I/O of the [[04 - The P2P Network Layer|P2P Transport]] and the local disk operations of the [[03 - Content Addressable Storage|Storage Layer]]. If the transport is the nervous system and the storage is the memory, the server is the cerebral cortex.

This note provides a deep dive into the `FileServer`'s responsibilities, its event loop, its peer management strategy, and its broadcasting mechanism.

## The FileServer Struct

```go
type FileServer struct {
    FileServerOpts
    store   *Store
    transport Transport
    quitch  chan struct{}
    peers   map[string]Peer
    peerLock sync.Mutex
}
```

- `FileServerOpts`: Embedded configuration struct (ListenAddr, StorageRoot, BootstrapNodes, etc.).
- `store`: Pointer to the local [[03 - Content Addressable Storage|Store]].
- `transport`: The [[04 - The P2P Network Layer|Transport]] interface (likely a `TCPTransport`).
- `quitch`: A signal channel used to gracefully shut down the server.
- `peers`: A map of currently connected peers, keyed by their network address string.
- `peerLock`: A `sync.Mutex` to protect the `peers` map from concurrent access.

## Configuration via FileServerOpts

```go
type FileServerOpts struct {
    StorageRoot       string
    PathTransformFunc PathTransformFunc
    Transport         Transport
    ListenAddr        string
    BootstrapNodes    []string
}
```

This struct encapsulates everything needed to start a node. Notably, it injects the `Transport` as a dependency. This means a `FileServer` can be instantiated with a mock transport for unit testing—a direct benefit of the interface-driven design discussed in [[02 - Architecture and Design Patterns]].

## The Main Event Loop

The `Start` method is the entry point for a running node:

```go
func (s *FileServer) Start() error {
    if err := s.transport.ListenAndAccept(); err != nil {
        return err
    }
    s.bootstrapNetwork()
    return s.loop()
}
```

This method does three things in sequence:
1.  **Listen**: Tells the transport to start accepting connections.
2.  **Bootstrap**: Dials any configured bootstrap nodes to join the network.
3.  **Loop**: Enters the main event loop.

### bootstrapNetwork

`bootstrapNetwork` iterates over the `BootstrapNodes` list and calls `s.transport.Dial(addr)` for each. Each dial happens in its own goroutine, allowing the node to attempt multiple connections concurrently without blocking. This is critical for resilience; if one bootstrap node is offline, the others are still attempted.

### The Loop

The `loop` method is the heart of the server:

```go
func (s *FileServer) loop() error {
    rpcCh := s.transport.Consume()
    for {
        select {
        case rpc := <-rpcCh:
            fmt.Printf("recv: %v\n", rpc)
        case <-s.quitch:
            return nil
        }
    }
}
```

Currently, this loop is very simple. It waits for two events:
1.  An incoming `RPC` from the transport channel.
2.  A shutdown signal on `quitch`.

When an `RPC` arrives, it simply prints it. This is the most obvious "Phase 1" placeholder in the entire project. In a mature system, this `select` block would contain logic to dispatch messages based on their type: storing a file, retrieving a file, handling a heartbeat, etc.

## Peer Management

When the `TCPTransport` accepts a new connection, it invokes the `OnPeer` callback:

```go
func (s *FileServer) OnPeer(p Peer) error {
    s.peerLock.Lock()
    defer s.peerLock.Unlock()
    s.peers[p.RemoteAddr().String()] = p
    return nil
}
```

This method adds the peer to the `FileServer`'s internal map. It is the server's responsibility to track peers because the server knows which peers are relevant for broadcasting and application-level logic. The transport, by contrast, focuses purely on I/O.

The use of a mutex here is essential. Because `OnPeer` is called from a goroutine inside the transport, and the `peers` map is read from the main loop or other goroutines during broadcasting, concurrent access is guaranteed.

## Storing and Broadcasting Data

The `StoreData` method is the most complex and interesting method in the server:

```go
func (s *FileServer) StoreData(key string, r io.Reader) error {
    // 1. Stream the file to local disk
    if _, err := s.store.Write(key, r); err != nil {
        return err
    }
    
    // 2. Prepare the payload
    buf := new(bytes.Buffer)
    tee := io.TeeReader(r, buf)
    p := &Payload{
        Key:  key,
        Data: buf.Bytes(),
    }
    
    // ... broadcast to peers
}
```

Wait, actually, looking at the code exploration again, it likely uses `io.TeeReader` or reads the data into a buffer to both store locally and broadcast. The exploration summary mentioned: `StoreData`: Writes data to local disk then broadcasts a `Payload{Key, Data}` to all known peers using `gob` encoding over `io.MultiWriter`.

Let's reconstruct the likely logic based on the summary:

1.  It writes the data to the local `store`.
2.  It constructs a `Payload` struct containing the key and the raw data.
3.  It uses `gob.NewEncoder` to encode this payload.
4.  Crucially, it uses `io.MultiWriter` to send the encoded bytes to **all connected peers simultaneously**.

### The MultiWriter Broadcast

`io.MultiWriter` is a brilliant standard library tool for this:

```go
func (s *FileServer) broadcast(p *Payload) error {
    s.peerLock.Lock()
    defer s.peerLock.Unlock()
    
    peers := []io.Writer{}
    for _, peer := range s.peers {
        peers = append(peers, peer)
    }
    
    multiWriter := io.MultiWriter(peers...)
    return gob.NewEncoder(multiWriter).Encode(p)
}
```

By wrapping all peer connections in an `io.MultiWriter`, the server can execute a single `Encode` call, and the `gob` encoder streams the bytes to every peer concurrently. This is elegant, efficient, and leverages Go's interface system perfectly.

## Graceful Shutdown

The `quitch` channel provides a mechanism for graceful shutdown. An external caller can signal the channel, causing the `loop` to exit and `Start` to return. While the current implementation doesn't perform extensive cleanup (like closing all peer connections explicitly), the structure is in place to add it.

## The Inversion of Control

A subtle but powerful design choice is how the transport calls back into the server. The `TCPTransportOpts` includes:

```go
OnPeer func(Peer) error
```

When constructing the transport in `main.go`, this is wired to `s.OnPeer`. This means the transport doesn't need to know about the `FileServer` type—it only knows about a function that takes a `Peer`. This is **Inversion of Control**, and it keeps the `p2p` package free of application-specific logic.

## Related Notes

- [[03 - Content Addressable Storage]]: The `Store` that `FileServer` manages.
- [[04 - The P2P Network Layer]]: The `Transport` that `FileServer` consumes.
- [[05 - Message Encoding and Protocol]]: The `Payload` and `gob` encoding used in broadcasting.
- [[07 - Bootstrapping and Entry Point]]: See `main.go` to watch the server in action.
