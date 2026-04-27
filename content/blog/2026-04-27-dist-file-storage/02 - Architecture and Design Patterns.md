---
title: "02 - Architecture and Design Patterns"
tags: [architecture, design-patterns, go, interfaces]
---

# 02 - Architecture and Design Patterns

One of the most impressive aspects of `dist_file_storage` is its disciplined use of Go idioms and design patterns. Despite being an early-stage project, it follows a clean, layered architecture that would not look out of place in a much larger codebase. This note explores the structural decisions that make the system extensible and testable.

## The Three Layers

The system is divided into three primary layers, each with a single, well-defined responsibility:

1.  **Storage Layer** (`storage.go`): Handles local disk I/O. It knows nothing about networks.
2.  **Network/Transport Layer** (`p2p/`): Handles TCP connections, peers, and raw message frames. It knows nothing about files.
3.  **Server/Orchestration Layer** (`server.go`): Wires storage and transport together. It manages peer connections and runs the main event loop.

This separation is crucial. It means you could theoretically swap out the TCP transport for a UDP or WebSocket implementation without touching the storage logic. Similarly, you could change the path hashing algorithm without affecting the network layer.

## Interface-Driven Design

Go's implicit interfaces are used heavily to enforce this decoupling.

### The Transport Interface

Defined in `p2p/transport.go`, the `Transport` interface abstracts everything a node needs to communicate:

```go
type Transport interface {
    ListenAndAccept() error
    Consume() <-chan RPC
    Close() error
    Dial(addr string) error
}
```

The `FileServer` in `server.go` holds a `Transport`, not a `TCPTransport`. This means the server logic is completely decoupled from TCP specifics. If you wanted to build a `UDPTransport` tomorrow, you could, as long as it satisfied this contract.

### The Peer Interface

Similarly, the `Peer` interface abstracts a network connection:

```go
type Peer interface {
    net.Conn
    Send([]byte) error
}
```

This allows the server to treat all connections uniformly, whether they are inbound or outbound.

### The Decoder Interface

In `p2p/encoding.go`, the `Decoder` interface allows pluggable message framing:

```go
type Decoder interface {
    Decode(io.Reader, *RPC) error
}
```

The project ships with a `GOBDecoder` and a `DefaultDecoder` (which reads a fixed 1028-byte buffer). This pattern anticipates future needs: perhaps a `ProtobufDecoder` or a `JSONDecoder` could be dropped in later.

## The Options Struct Pattern

Every major component is configured via an "Options" struct:

- `StoreOpts` for the storage layer.
- `TCPTransportOpts` for the TCP transport.
- `FileServerOpts` for the server.

This is a clean alternative to long constructor parameter lists. It allows for optional configuration, sensible defaults, and forward-compatible APIs. For example, `StoreOpts` lets you inject a custom `PathTransformFunc` without changing the `NewStore` signature.

## Strategy Pattern via Functions

Several behaviors are injected as functions rather than hardcoded:

- `PathTransformFunc`: Defines how a key maps to a file path.
- `HandshakeFunc`: Defines what happens when a peer connects.

This is the **Strategy Pattern** in action. The default path transform is an identity function, but the system primarily uses `CASPathTransformFunc`, which SHA-1 hashes the key. This flexibility is powerful and keeps the core structs agnostic to specific algorithms.

## Concurrency Patterns

The system is goroutine-heavy, which is idiomatic for Go networking:

- `ListenAndAccept` spawns a goroutine for the accept loop.
- `startAcceptLoop` spawns a goroutine for each accepted connection (`handleConn`).
- `bootstrapNetwork` spawns goroutines to dial peers concurrently.

To protect shared state, it uses `sync.RWMutex` in the `Store` and `sync.Mutex` (`peerLock`) in the `FileServer` for the peers map. The `RPC` messages are passed between goroutines via Go channels (`Consume() <-chan RPC`), which is the canonical way to share memory by communicating.

## Why This Matters

These patterns aren't just academic. They solve real problems:

- **Testability**: You can mock the `Transport` interface to test `FileServer` logic in isolation.
- **Extensibility**: Adding UDP support is a matter of implementing an interface, not refactoring the world.
- **Clarity**: A new contributor can look at `server.go` and understand the high-level flow without getting lost in TCP socket details.

## Related Notes

- [[03 - Content Addressable Storage]]: See how the storage layer implements these patterns.
- [[04 - The P2P Network Layer]]: See the transport interfaces in action.
- [[06 - Server Orchestration]]: See how the server wires it all together.
