---
title: "09 - Current State and Future Roadmap"
tags: [roadmap, future, distributed-systems, phase-1, missing-features]
---

# 09 - Current State and Future Roadmap

Every ambitious project begins as a scaffold. `dist_file_storage` is no exception. While its architecture is thoughtfully designed, it is crucial to be honest about where it stands today and where it could go tomorrow. This note provides a transparent assessment of the current implementation and a speculative roadmap for turning this scaffold into a production-adjacent distributed storage system.

## Current State: Phase 1 Scaffold

The project has successfully laid the groundwork. The following components are functional and well-architected:

- **Content-Addressable Storage**: The `Store` and `CASPathTransformFunc` work correctly. Files can be written, read, checked, and deleted using SHA-1 based paths.
- **TCP Transport**: Nodes can listen for connections and accept raw byte payloads.
- **Peer Connection**: Nodes can dial bootstrap peers and establish TCP connections.
- **Broadcasting**: The `FileServer` can broadcast a `gob`-encoded `Payload` to all connected peers using `io.MultiWriter`.
- **Test Harness**: `main.go` demonstrates two local nodes interacting, and a `Makefile` streamlines the build.

These are non-trivial achievements. The project has a heartbeat.

## What's Missing

The README and codebase explicitly acknowledge several gaps. Here is a consolidated view of the eight major areas needing work:

### 1. Real Request/Response Protocol
The `FileServer`'s `loop` currently just prints incoming `RPC` messages. There is no logic to parse a message type (e.g., `STORE`, `GET`, `LIST`) and act on it. The system cannot handle a peer asking for a file it doesn't have.

### 2. Outbound Connection Management
While `Dial` exists, connections are not robustly managed. There is no reconnection logic, no heartbeat/ping mechanism, and no timeout handling. If a peer disconnects, the server may not notice until it tries to broadcast.

### 3. GOBDecoder Integration
`GOBDecoder` is implemented but sits unused. The transport defaults to `DefaultDecoder`, which reads fixed 1028-byte buffers. The broadcast path uses `gob`, but the general receive path does not. The wire protocol needs to be unified.

### 4. Message Loop Logic
The `loop` method in `server.go` is a placeholder. It needs a dispatcher that can:
- Handle incoming store requests.
- Respond to file retrieval requests by reading from the `Store` and streaming the data back.
- Manage peer health and propagate disconnections.

### 5. Replication and Sharding
Currently, broadcasting sends the full file to every peer. This is not scalable. A real system needs:
- **Replication**: Store `N` copies of a file across the network for redundancy.
- **Sharding**: Split large files into chunks and distribute them.

### 6. Consensus and Metadata
There is no shared state. If Node 1 has a file and Node 2 doesn't, Node 2 has no way of knowing that Node 1 has it without asking every node. A distributed hash table (DHT) like Kademlia would solve this.

### 7. Deployment and Operations
The project is purely a dev harness. There is no:
- Command-line interface (CLI).
- Configuration file support.
- Docker or containerization.
- Logging framework (just `fmt.Printf`).
- Metrics or monitoring.

### 8. Security
There is no encryption, no authentication, and no verification that received data matches its hash. A malicious peer could send garbage data, and the current system would write it to disk.

## The Roadmap: A Path Forward

How would one evolve this project? Here is a speculative, phase-by-phase roadmap.

### Phase 2: The Protocol
- **Define a message format**: Implement a length-prefixed framing protocol.
- **Message types**: `STORE`, `GET`, `DELETE`, `HANDSHAKE`, `HEARTBEAT`.
- **Request IDs**: Correlate responses with requests.
- **Dispatcher**: Replace the print statement in `loop` with a real message router.
- **GOB or Protobuf**: Standardize on `gob` for Go-only or migrate to Protocol Buffers for cross-language support.

### Phase 3: File Retrieval
- Implement `GetData(key string)` on the `FileServer`.
- When a `GET` request arrives, look up the key in the local `Store`.
- If found, stream the file back to the requester.
- If not found, forward the request to known peers (recursive lookup).

### Phase 4: Resilient Networking
- Add a heartbeat/ping mechanism to detect dead peers.
- Implement exponential backoff reconnection for bootstrap nodes.
- Add connection timeouts and graceful shutdown of peer connections.
- Use a context-based cancellation strategy.

### Phase 5: Distributed Hash Table (DHT)
- Integrate a Kademlia-style DHT or a simplified chord ring.
- Allow nodes to find which peer holds a given hash without broadcasting to everyone.
- This is the step that transforms the system from a broadcast mesh into a scalable network.

### Phase 6: Chunking and Erasure Coding
- Split large files into fixed-size blocks (e.g., 256KB).
- Use a Merkle tree to verify block integrity.
- Implement Reed-Solomon erasure coding for redundancy without full replication.

### Phase 7: Production Hardening
- Build a CLI with `cobra` or `urfave/cli`.
- Add structured logging with `zap` or `logrus`.
- Containerize with Docker and provide a `docker-compose.yml` for local clusters.
- Add Prometheus metrics.
- Implement TLS for peer connections.

## Why This Project Matters

Despite its incompleteness, `dist_file_storage` is an excellent educational artifact. It demonstrates:
- How to layer a distributed system.
- How to use Go interfaces for testability.
- How to implement CAS on a local filesystem.
- How to structure a P2P network without massive frameworks.

Every missing feature is an opportunity to learn. The gap between Phase 1 and Phase 7 is exactly the gap between a student project and a system like IPFS—and that gap is precisely what makes this codebase such a valuable starting point.

## Related Notes

- [[01 - Project Overview]]: The original goals of the project.
- [[02 - Architecture and Design Patterns]]: The solid foundation that makes this roadmap feasible.
- [[06 - Server Orchestration]]: The layer where most of the Phase 2 work will happen.
- [[07 - Bootstrapping and Entry Point]]: Where Phase 7 hardening will begin.
