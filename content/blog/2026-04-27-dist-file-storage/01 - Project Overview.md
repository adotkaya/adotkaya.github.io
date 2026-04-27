---
title: "01 - Project Overview"
tags: [overview, go, p2p, distributed-systems]
---

# 01 - Project Overview

`dist_file_storage` is a nascent, peer-to-peer distributed file storage system written in **Go 1.25**. At its core, it attempts to answer a fundamental question: *how do you store files across a network of untrusted or semi-trusted nodes in a way that is robust, content-addressable, and decentralized?*

While the project is currently in an early **scaffold/Phase 1** state, its architecture is deliberately designed to evolve into a fully functional system reminiscent of early BitTorrent or IPFS-like networks.

## The Core Idea

In a traditional client-server model, you upload a file to a central server. In `dist_file_storage`, a file is written to a node's local disk and then propagated across the network to all connected peers. Files are located not by a path or a URL, but by the **hash of their content**. This is the principle of [[03 - Content Addressable Storage|Content Addressable Storage (CAS)]].

## Why This Project Exists

Building distributed systems is hard. There are countless subtle failure modes: network partitions, Byzantine peers, data corruption, and the eternal difficulty of maintaining consensus. This project exists as a ground-up exploration of these challenges. It strips away the complexity of a production system to focus on the foundational layers:

1.  **How do you reliably store a file on disk using a content hash?**
2.  **How do you establish and maintain TCP connections between peers?**
3.  **How do you broadcast data to the network once it's stored locally?**

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25 |
| Networking | Standard Library (`net`) |
| Serialization | `encoding/gob` (for broadcasts), raw byte buffers (for transport) |
| Testing | `testing`, `github.com/stretchr/testify` |
| Build | `Makefile` |

The deliberate choice to use the Go standard library for networking is significant. By avoiding heavy frameworks like gRPC or libp2p at this stage, the project maintains full control over the wire protocol and connection lifecycle. This is an educational and architectural choice that prioritizes understanding over convenience.

## Current Capabilities

As of the latest commit, the system can:
- Store files locally using a SHA-1 based path transformation.
- Listen for incoming TCP connections from other peers.
- Dial bootstrap nodes on startup.
- Broadcast a `Payload` struct (encoded via `gob`) to all connected peers.
- Run a local dev harness with two interacting nodes.

## What It Is Not (Yet)

It is important to set expectations. This is not a production-ready system. There is no request/response protocol for retrieving files over the network, no replication strategy, no sharding, and no consensus mechanism. The [[09 - Current State and Future Roadmap|Current State and Future Roadmap]] details these gaps honestly.

## Where to Go Next

- To understand the high-level design, read [[02 - Architecture and Design Patterns]].
- To jump straight into the disk layer, read [[03 - Content Addressable Storage]].
- To see how nodes talk to each other, read [[04 - The P2P Network Layer]].
