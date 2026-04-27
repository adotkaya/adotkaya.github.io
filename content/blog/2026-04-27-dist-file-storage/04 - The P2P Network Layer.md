---
title: "04 - The P2P Network Layer"
tags: [p2p, networking, tcp, go, transport]
---

# 04 - The P2P Network Layer

If the storage layer is the foundation of `dist_file_storage`, then the P2P network layer is its nervous system. Located in the `p2p/` package, this code is responsible for the most complex and error-prone part of any distributed system: **networking**.

This note explores the abstractions, implementations, and concurrency model that allow nodes to discover, connect to, and communicate with each other over raw TCP.

## Core Abstractions: Transport and Peer

The first thing you notice when opening `p2p/transport.go` is the commitment to interfaces.

### The Transport Interface

```go
type Transport interface {
    ListenAndAccept() error
    Consume() <-chan RPC
    Close() error
    Dial(addr string) error
}
```

This interface is the contract that the `FileServer` (see [[06 - Server Orchestration]]) uses to interact with the network. It is intentionally minimal:

- `ListenAndAccept()`: Start listening for inbound connections.
- `Consume()`: Return a read-only channel of incoming messages (`RPC`).
- `Close()`: Shut down the transport.
- `Dial(addr string)`: Connect to a remote peer.

By programming against this interface, the server remains agnostic to whether the underlying protocol is TCP, UDP, WebSockets, or even an in-memory channel for testing.

### The Peer Interface

```go
type Peer interface {
    net.Conn
    Send([]byte) error
}
```

A `Peer` is essentially a network connection with an added `Send` method. It wraps `net.Conn` to provide a slightly higher-level API for sending raw bytes. The `TCPPeer` implementation also tracks whether the connection was initiated outbound, which can be useful for debugging and topology management.

## TCP Transport Implementation

The concrete implementation, `TCPTransport`, lives in `p2p/tcp_transport.go`.

### Struct Definition

```go
type TCPTransport struct {
    listenAddress string
    listener      net.Listener
    mu            sync.RWMutex
    peers         map[net.Addr]Peer
    handshakeFunc HandshakeFunc
    decoder       Decoder
    onPeer        func(Peer) error
}
```

Key fields:
- `listenAddress`: The local bind address (e.g., `:3000`).
- `listener`: The active TCP listener.
- `peers`: A map of connected peers. *Note: As of the current implementation, this map exists in the transport but is not populated; the `FileServer` maintains its own peer map via the `onPeer` callback.*
- `handshakeFunc`: A function to run upon new connections.
- `decoder`: The pluggable message decoder.
- `onPeer`: A callback invoked when a new peer is accepted.

### The Lifecycle of a Connection

The `TCPTransport` manages connections through a hierarchy of goroutines:

1.  **ListenAndAccept**: Called to start the transport. It creates a `net.Listener` and spawns the `startAcceptLoop` goroutine.
2.  **startAcceptLoop**: Runs an infinite loop calling `listener.Accept()`. For every incoming connection, it spawns a `handleConn` goroutine.
3.  **handleConn**: 
    - Performs the optional `handshakeFunc`.
    - Calls the `onPeer` callback to notify the server.
    - Enters a read loop, using the `decoder` to parse messages into `RPC` structs.
    - Pushes each `RPC` onto an internal channel consumed by `Consume()`.

This is a classic Go network server pattern. The use of goroutines means the transport can handle thousands of concurrent connections, limited only by system resources.

### TCPPeer

```go
type TCPPeer struct {
    conn     net.Conn
    outbound bool
}
```

`TCPPeer` wraps a `net.Conn`. The `outbound` boolean distinguishes between connections we dialed (outbound) and connections that dialed us (inbound). This is a small but important piece of metadata for building a symmetric P2P network where every node is both a client and a server.

## Bootstrapping the Network

A node isn't very useful if it sits alone. The `TCPTransport` implements `Dial(addr string)`, which initiates an outbound TCP connection. The `FileServer` uses this in its `bootstrapNetwork` method, dialing a list of "bootstrap nodes" provided in its configuration.

This is how a new node joins an existing swarm. It knows about one or more well-known addresses, connects to them, and (in a future version) discovers more peers through them.

## The RPC Message Envelope

Before a message can be decoded, it needs a structure to hold it. `p2p/message.go` defines:

```go
type RPC struct {
    From    net.Addr
    Payload []byte
}
```

This is a minimal envelope. It doesn't prescribe what the payload contains—that's up to the application layer. It simply says: "These bytes came from this address." The `FileServer`'s main loop reads `RPC` structs from the transport channel and acts on them.

## Pluggable Decoding

The transport doesn't assume a fixed wire format. It uses the `Decoder` interface:

```go
type Decoder interface {
    Decode(io.Reader, *RPC) error
}
```

Currently, two decoders exist:
- `DefaultDecoder`: Reads a fixed 1028-byte buffer from the connection. Simple, but inefficient for small messages and limiting for large ones.
- `GOBDecoder`: Uses Go's `encoding/gob`. More structured, but currently not wired into the transport by default.

This design anticipates a future where the protocol might use length-prefixed framing, Protocol Buffers, or JSON. For more details, see [[05 - Message Encoding and Protocol]].

## Concurrency and Error Handling

Networking code is notoriously difficult because of concurrency and partial failures. The `TCPTransport` handles this by:

- Never blocking the accept loop. `handleConn` runs in its own goroutine, so a slow peer cannot prevent the node from accepting new connections.
- Using channels to communicate with the main server loop. This avoids shared mutable state between the transport and the server.
- Letting the server manage the peer lifecycle. The transport focuses on I/O; the server decides what to do with peers.

## Related Notes

- [[02 - Architecture and Design Patterns]]: Understand why interfaces are used here.
- [[05 - Message Encoding and Protocol]]: Dive deeper into `RPC`, `Decoder`, and `HandshakeFunc`.
- [[06 - Server Orchestration]]: See how the server consumes the transport's channel and manages peers.
