---
title: "05 - Message Encoding and Protocol"
tags: [protocol, encoding, gob, rpc, serialization]
---

# 05 - Message Encoding and Protocol

A network is just a pipe for bytes. The real challenge of distributed systems is agreeing on what those bytes mean. In `dist_file_storage`, the project takes a pragmatic, layered approach to message encoding: a minimal envelope for transport and a pluggable decoder for interpretation.

This note examines the message structures, the encoding strategies, and the handshake mechanism that define the wire protocol.

## The RPC Envelope

As seen in `p2p/message.go`, the fundamental unit of communication is the `RPC` struct:

```go
type RPC struct {
    From    net.Addr
    Payload []byte
}
```

This is intentionally bare-bones. `From` is set by the transport layer to identify the sender. `Payload` is an opaque byte slice. The transport layer does not attempt to parse the payload; it merely delivers it. This separation of concerns is powerful because it allows the application layer (the `FileServer`) to evolve its protocol without changing the transport code.

## The Decoder Interface

How does the transport turn a raw stream of bytes into an `RPC`? It delegates this to a `Decoder`:

```go
type Decoder interface {
    Decode(io.Reader, *RPC) error
}
```

This interface is the boundary between raw I/O and structured data. The project provides two implementations.

### DefaultDecoder

`DefaultDecoder` is the simplest possible decoder:

```go
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
    buf := make([]byte, 1028)
    n, err := r.Read(buf)
    if err != nil {
        return err
    }
    msg.Payload = buf[:n]
    return nil
}
```

It allocates a 1028-byte buffer and reads whatever is available on the connection into it. This approach has trade-offs:

**Pros:**
- Zero dependencies. No JSON library, no protobuf compiler, no `gob` registration.
- Extremely fast for prototyping.
- Works for any byte stream.

**Cons:**
- Fixed buffer size. Messages larger than 1028 bytes are truncated.
- No framing. If two small messages arrive quickly, they might be read into the same buffer (TCP is a stream protocol, not a message protocol).
- Wasteful for tiny messages.

`DefaultDecoder` is wired into the `TCPTransport` by default in the current `main.go` setup. It is a "good enough" solution for Phase 1.

### GOBDecoder

`GOBDecoder` uses Go's built-in `encoding/gob` package:

```go
type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
    return gob.NewDecoder(r).Decode(msg)
}
```

`gob` is Go's native binary serialization format. It is self-describing, efficient, and requires no external schema definitions. However, it is Go-specific.

**Pros:**
- Handles complex structs natively.
- Automatic framing (the decoder knows when one object ends and another begins).
- Type-safe.

**Cons:**
- Go-only. If you wanted to write a client in Rust or Python, `gob` would be a poor choice.
- Requires registration of types if using interfaces.

Interestingly, `GOBDecoder` is implemented but currently **unused** in the main transport loop. It exists as a forward-looking hook. If the project moves to a structured request/response protocol, `gob` might become the primary format, or it might be replaced by Protocol Buffers for cross-language compatibility.

## The Handshake Mechanism

Before two peers start exchanging data, they might want to verify each other or exchange metadata. `p2p/handshake.go` defines:

```go
type HandshakeFunc func(Peer) error
```

And provides a no-op default:

```go
func NOPHandshakeFunc(Peer) error { return nil }
```

This is another example of the Strategy Pattern. The `TCPTransport` accepts a `HandshakeFunc` in its options. Currently, it does nothing, but the hook is there for future use. Potential handshake logic could include:

- **Version Negotiation**: Ensuring both peers speak the same protocol version.
- **Authentication**: Exchanging tokens or certificates.
- **Capability Advertisement**: Telling the peer what features you support (e.g., "I can store files" vs. "I am just a relay").

## The Application-Level Payload

While the transport deals in `RPC` envelopes, the application layer defines what goes inside `Payload`. In `server.go`, the `FileServer` uses a `Payload` struct for broadcasting:

```go
type Payload struct {
    Key  string
    Data []byte
}
```

When `StoreData` is called, it writes this struct to all peers using `gob.NewEncoder(io.MultiWriter(...)).Encode(payload)`. This is a fascinating inversion: the transport uses `DefaultDecoder`, but the application-level broadcast uses `gob`. This is a sign of a project in transition. The broadcast mechanism is ahead of the general transport decoding.

## What's Missing: A Real Protocol

The current encoding setup is fragmented. There is no unified request/response protocol. A complete protocol would likely include:

1.  **Length-Prefixed Framing**: Every message is preceded by its length in bytes. This solves the "message boundary" problem of TCP.
2.  **Message Types**: An enum or string indicating whether the payload is a `STORE`, `GET`, `DELETE`, or `HANDSHAKE` message.
3.  **Request IDs**: So that responses can be correlated with requests.
4.  **Checksums**: To detect data corruption in transit.

These are standard features of protocols like Redis RESP, HTTP/2 frames, or Bitcoin's P2P protocol. Implementing one of these would be a major milestone for the project.

## Related Notes

- [[04 - The P2P Network Layer]]: See where `Decoder` and `HandshakeFunc` are used in `TCPTransport`.
- [[06 - Server Orchestration]]: See how the server uses `Payload` and `gob` for broadcasting.
- [[09 - Current State and Future Roadmap]]: Understand why the protocol is still in flux.
