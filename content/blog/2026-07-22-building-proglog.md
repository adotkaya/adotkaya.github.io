---
title: "Building Proglog: What the Book Showed Me, and What I Had to See for Myself"
date: "2026-07-22"
slug: "building-proglog"
---

A few weeks ago I picked up *Distributed Services with Go* by Travis Jeffery with a simple goal: build something that actually runs across multiple machines. I wanted to stop reading about distributed systems and start *feeling* them—network partitions, consensus, replication, the whole stack. The book delivered a working system. But what it didn't deliver was a mental map. This post is that map.

## The Premise

Proglog is an append-only commit log service. Think of it as a tiny Kafka or a educational NATS JetStream. It stores records in segmented files, exposes a gRPC API, replicates across nodes, and elects leaders via Raft. I built it commit-by-commit, and the Git history tells the story: basic log → gRPC → segmented storage → replication → Raft → client-side load balancing.

The book's approach is practical: "Here is the code. Type it. Run the tests." That gets you a green test suite after a bit of work (Author gets a bit lost on the structure, mess on code structure and lots of typos etc...oh and the incompatible versions). But after fixing and gluing the code together, I found myself staring at passing tests wondering *why* the system was shaped the way it was. So I wrote an `ARCHITECTURE.md`—not for the repo, but for myself. This post is the human-readable version.

## The Storage Engine

The heart of proglog is a layered storage stack:

```
Log (manager)
  └── []*segment
        └── segment (active)
              ├── Store  (buffered file I/O)
              └── Index  (memory-mapped file)
```

### Store: Length-Prefixing

The `Store` is a buffered file writer. Records are written with an 8-byte length prefix followed by the raw payload:

```go
func (s *Store) Append(p []byte) (n uint64, pos uint64, err error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    pos = s.size
    // Write 8-byte length + payload to buffered writer
    if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
        return 0, 0, err
    }
    if _, err := s.buf.Write(p); err != nil {
        return 0, 0, err
    }
    s.size += uint64(lenWidth + len(p))
    return uint64(lenWidth + len(p)), pos, nil
}
```

This simple decision—length-prefixing—is what makes random access possible. You don't scan. You jump to a position, read 8 bytes to know the payload size, then read exactly that many bytes. The buffer flushes to disk only on `Read`, `ReadAt`, or `Close`, batching syscalls.

### Index: Relative Offsets and mmap

The `Index` maps offsets to byte positions in the Store. It uses memory-mapped files via `gommap` for fast random access. Each index entry is exactly 12 bytes:

```
| 4 bytes: uint32 relative offset | 8 bytes: uint64 position |
```

Why *relative* offsets? If the first record in a segment has absolute offset 16, its relative offset is 0. The next is 1, then 2. At scale, using `uint32` instead of `uint64` saves 4 bytes per entry. When you are indexing billions of records, that matters ofc.

```go
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
    if in == -1 {
        out = uint32((i.size / entWidth) - 1)
    } else {
        out = uint32(in)
    }
    pos = uint64(out) * entWidth
    if i.size < pos+entWidth {
        return 0, 0, io.EOF
    }
    out = enc.Uint32(i.mmap[pos : pos+offWidth])
    pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
    return out, pos, nil
}
```

On `Close`, the index `Sync`s the mmap region, `Sync`s the file, then truncates to the actual written size. Clean and durable.

### Segment Rotation

A `Segment` binds one Store to one Index. When either hits its configured max size, the `Log` manager creates a new segment with a `baseOffset` equal to the next expected absolute offset. Old segments stay on disk until truncated. This keeps files bounded and makes compaction simple.

## gRPC API and Security

I didn't worked with gRPC before this project, at least this book with its errors made me realize that lots of time should be spent understanding it, not just copying code or prompting to an AI, LLM's and moving with it. You would eventually crash into a wall of errors. Better start learning it!

The service exposes four RPCs: unary `Produce`/`Consume`, and streaming `ProduceStream`/`ConsumeStream`. The generated protobuf is straightforward, but the server setup is where production concerns appear.

The server chains four interceptors:

1. `grpc_ctxtags` — extracts tags for logging context
2. `grpc_zap` — structured request logging with nanosecond durations
3. `grpc_auth` — extracts the client identity from the TLS handshake
4. `ocgrpc` — OpenCensus metrics and traces

### mTLS and Authorization

Transport security uses mutual TLS: the server presents a certificate, and the client must present one too. The auth interceptor pulls the Common Name from the verified certificate chain and stores it in the request context:

```go
func authenticate(ctx context.Context) (context.Context, error) {
    peer, _ := peer.FromContext(ctx)
    if peer.AuthInfo == nil {
        return ctx, nil
    }
    tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
    subject := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
    return context.WithValue(ctx, subjectContextKey{}, subject), nil
}
```

But identity is not permission. For that, proglog uses Casbin. The policy file is tiny:

```csv
p, root, *, produce
p, root, *, consume
```

A client with CN `root` can do anything. A client with CN `nobody` cannot. The tests generate both certificates to prove the boundary works. Separating transport auth (mTLS) from policy auth (Casbin) is a clean design decision I appreciated once I stopped to think about it.

## Serf Discovery and Replication

Before Raft, the book introduces replication via HashiCorp Serf, a gossip-based membership protocol. When a node joins the cluster, Serf broadcasts the event. A `Membership` component listens and calls a `Handler` interface:

```go
type Handler interface {
    Join(name, addr string) error
    Leave(name string) error
}
```

The `Replicator` implements this interface. On `Join`, it spawns a goroutine that dials the peer and opens a `ConsumeStream` starting at offset 0. Every record received is re-produced into the local log:

```go
func (r *Replicator) replicate(addr string, leave chan struct{}) {
    conn, _ := grpc.Dial(addr, r.DialOptions...)
    client := api.NewLogClient(conn)
    stream, _ := client.ConsumeStream(ctx, &api.ConsumeRequest{Offset: 0})
    for {
        select {
        case <-r.close:
            return
        case <-leave:
            return
        case record := <-records:
            r.LocalServer.Produce(ctx, &api.ProduceRequest{Record: record})
        }
    }
}
```

This works for simple topologies, but it is not consensus. If two nodes produce different records at the same offset, the system has no mechanism to resolve the conflict. That realization leads to the next phase.

## Raft and the Agent

Raft solves the conflict problem by ensuring only one node—the leader—accepts writes. All other nodes replicate the leader's log. Proglog wraps the local `Log` inside a `DistributedLog` that embeds HashiCorp's Raft implementation.

### The Finite State Machine

Next, Raft doesn't know what a "record" is. It knows log entries and commands. Proglog bridges this gap with an FSM (Finite State Machine, saddly not Fatih Sultan Mehmet, my precious...):

```go
type fsm struct {
    log *Log
}

func (f *fsm) Apply(record *raft.Log) interface{} {
    reqType := RequestType(record.Data[0])
    switch reqType {
    case AppendRequestType:
        return f.applyAppend(record.Data[1:])
    }
    return nil
}
```

When the leader receives a `Produce` request, it serializes the request type and protobuf payload, then calls `raft.Apply`. Once a majority of nodes append the command to their Raft log, the FSM's `Apply` method runs on each node, appending the record to the local storage engine. Reads bypass Raft entirely—they go straight to the local `Log`.

### LogStore Adapter

Raft needs a `LogStore` interface. Rather than building a new database, proglog adapts the existing `Log` type:

```go
func (l *logStore) StoreLogs(records []*raft.Log) error {
    for _, record := range records {
        _, err := l.Append(&api.Record{
            Value: record.Data,
            Term:  record.Term,
            Type:  uint32(record.Type),
        })
        if err != nil {
            return err
        }
    }
    return nil
}
```

The storage engine's offset becomes Raft's index. The same segmented files serve both the user-facing API and the consensus protocol.

### Connection Multiplexing with cmux

Here is where I stopped and stared for a while. Raft nodes need to talk to each other. gRPC clients need to talk to the server. The book uses a single TCP port for both, which sounds impossible until you see `cmux` in action.

`cmux` inspects the first bytes of a connection to decide where to route it. Raft connections write a single-byte header before the TLS handshake:

```go
func (s *StreamLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
    conn, _ := dialer.Dial("tcp", string(addr))
    _, err := conn.Write([]byte{byte(RaftRPC)})
    if err != nil {
        return nil, err
    }
    if s.peerTLSConfig != nil {
        conn = tls.Client(conn, s.peerTLSConfig)
    }
    return conn, err
}
```

On the server side, `cmux` matches this byte and routes the connection to Raft. Everything else goes to gRPC. The `Agent` orchestrates this entire lifecycle: mux setup → log with Raft → gRPC server → Serf membership, all wired together and gracefully shut down in reverse order.

## Client-Side Load Balancing

The final layer was the most surprising. Instead of a centralized load balancer, proglog implements custom gRPC client-side balancing. A `Resolver` queries the cluster via `GetServers` and returns addresses annotated with a leader flag:

```go
func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {
    client := api.NewLogClient(r.resolverConn)
    res, _ := client.GetServers(ctx, &api.GetServersRequest{})
    var addrs []resolver.Address
    for _, server := range res.Servers {
        addrs = append(addrs, resolver.Address{
            Addr: server.RpcAddr,
            Attributes: attributes.New("is_leader", server.IsLeader),
        })
    }
    r.clientConn.UpdateState(resolver.State{Addresses: addrs})
}
```

The `Picker` then routes `Produce` calls to the leader and round-robins `Consume` calls across followers:

```go
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
    var result balancer.PickResult
    if info.FullMethodName == "/log.v1.Log/Produce" {
        result.SubConn = p.leader
    } else if len(p.followers) > 0 {
        offset := atomic.AddUint64(&p.current, 1)
        result.SubConn = p.followers[offset%uint64(len(p.followers))]
    }
    return result, nil
}
```

This is elegant. The client *knows* the topology and makes intelligent routing decisions without a proxy in the path.

## What I Actually Learned

The book got me to a passing test suite. But the real learning happened when I wrote `ARCHITECTURE.md` and traced every data flow by hand. Here is what stuck:

- **Relative offsets matter.** A 4-byte vs 8-byte decision is not premature optimization when you are designing for scale.
- **Flush discipline matters.** Buffered writes are fast, but if you don't flush before reading, you read stale data. The store flushes on `Read`, `ReadAt`, and `Close`—a careful contract.
- **Raft is a state machine synchronizer, not a database.** Once I saw the FSM bridge, Raft stopped feeling like magic and started feeling like a replicated command log.
- **cmux is a cheat code.** One port, two protocols, zero proxies.
- **Client-side routing beats proxies for simple topologies.** When the client can ask "who is the leader?" and route accordingly, you remove a hop.


## Final Thoughts about the Book

The book is a *great engineer's build log*. And a *bad author's manifesto*. It shows you what to type. But distributed systems are not learned by typing. They are learned by tracing data through every layer and asking "*what if this fails?*" at every step. That is the work the book couldn't do for me.

To my criticism; you can say that "*Well do it yourself! Explore it!*" and you are probably right. My response to that, i think in *BIG 20th century*, this great project is not worth as a book. I would rather listen his thoughts and why he build it like that, what's are important etc.. in a video or podcast... Call me spoiled but i could've just copy it from Github, run it locally and observe it if i wanted to do that. (You can't btw, you should glue everything together to make it work)

I think a book either should teach you a thing, or push you to a critical thinking and researching process. This book does not fulfill either of those criteria. It just jumps jumps jumps.

And with this paragraph i just it realized that it has pushed me into a critical thinking and researching process with its incomplete content because i want to learn how to build distributed systems and make them stay alive. So, either way, thanks to Travis Jeffrey for his book!
