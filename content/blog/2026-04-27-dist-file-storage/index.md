---
title: "Building a Distributed File Storage System in Go"
date: "2026-04-27"
slug: "dist-file-storage"
tags: [distributed-systems, go, p2p, storage, project-deep-dive]
---

# Building a Distributed File Storage System in Go

Welcome to this comprehensive deep dive into `dist_file_storage`, an early-stage distributed file storage system written in Go. This project represents a fascinating journey into the core concepts of peer-to-peer networking, content-addressable storage, and layered system architecture.

This is not a tutorial, but rather a long-form exploration of the codebase, the design decisions, and the architectural patterns that underpin the system. Because this is an Obsidian vault, you can navigate through the ideas using the **Graph View** to see how concepts like [[03 - Content Addressable Storage|Content Addressable Storage]], [[04 - The P2P Network Layer|P2P Networking]], and [[06 - Server Orchestration|Server Orchestration]] connect.

## What You Will Find Here

- **A complete architectural breakdown** of a Go-based P2P storage system.
- **Deep dives into specific layers**: from disk I/O to TCP networking.
- **Design pattern analysis**: why interfaces, options structs, and callbacks are used.
- **An honest assessment** of the current state and a roadmap for the future.

## Start Here

If you are new to the project, begin with [[01 - Project Overview]] to understand the "what" and the "why". From there, you can follow the logical flow through the architecture or jump to any specific layer that interests you.

## The Structure of This Vault

| Note | Focus |
|------|-------|
| [[01 - Project Overview]] | Goals, tech stack, motivation |
| [[02 - Architecture and Design Patterns]] | High-level system design |
| [[03 - Content Addressable Storage]] | The disk-based storage layer (`storage.go`) |
| [[04 - The P2P Network Layer]] | TCP transport and peer abstractions (`p2p/`) |
| [[05 - Message Encoding and Protocol]] | RPC messages, decoders, and handshakes |
| [[06 - Server Orchestration]] | The `FileServer` brain (`server.go`) |
| [[07 - Bootstrapping and Entry Point]] | Running the system (`main.go`, `Makefile`) |
| [[08 - Testing Strategy]] | How the project is tested |
| [[09 - Current State and Future Roadmap]] | What's done and what's next |

---

*This vault is designed to be read linearly or explored as a graph. Every major concept is linked to its related files so you can trace dependencies across the system.*
