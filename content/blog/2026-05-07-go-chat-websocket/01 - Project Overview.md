---
title: "01 — Project Overview"
date: 2026-05-07
tags: [go, chat, websockets, jwt, postgresql, overview]
---

# 01 — Project Overview

`go-chat` is a real-time chat application with a terminal-style frontend, WebSocket-based messaging, JWT authentication, and PostgreSQL persistence. At its core, it attempts to answer a practical question: *how do you build a production-grade real-time system in Go without reaching for heavy frameworks or managed services?*

This project exists as a deliberate exercise in foundational backend engineering. It strips away the complexity of a production chat platform like Slack or Discord to focus on the layers that matter: stateful connections, authenticated APIs, persistent storage, and clean architecture.

## The Core Idea

In a traditional REST API, the server waits for the client to ask. In `go-chat`, the server pushes. When a user sends a message, every other user in the same room sees it instantly — no polling, no refreshing, no latency. This is enabled by WebSockets, but the real engineering is in how those connections are managed, secured, and scaled.

The application also demonstrates full-stack ownership: a Go backend, a vanilla JavaScript frontend, a PostgreSQL schema, Docker containerization, and Makefile-driven workflows. Every layer is intentional and explainable.

## Why This Project Exists

Real-time systems are a common interview topic, but building one from scratch teaches you things tutorials skip:

1.  **How do you authenticate a WebSocket connection?** HTTP headers during the upgrade, or query parameters?
2.  **How do you broadcast to 500 clients without blocking the sender?** Goroutines, channels, and non-blocking sends.
3.  **How do you prevent goroutine leaks when a client disappears?** Ping/pong, read deadlines, and deferred cleanup.
4.  **How do you structure a Go project so it's testable without a database?** Interfaces, dependency injection, and the repository pattern.

This project exists to explore these questions hands-on.

## Tech Stack

| Component | Technology | Why |
|-----------|------------|-----|
| Language | Go 1.24+ | Standard library richness, goroutines, compiled performance |
| Router | Chi (v5) | Lightweight, `context.Context` native, clean middleware chain |
| Database | PostgreSQL | ACID transactions, JSON support, robust concurrency |
| Driver | `lib/pq` | Mature, battle-tested PostgreSQL driver for Go |
| WebSockets | Gorilla WebSocket | De facto standard, excellent examples, production-hardened |
| Auth | JWT (golang-jwt) + bcrypt | Stateless tokens, adaptive password hashing |
| Frontend | Vanilla JS + CSS | No build step, no framework lock-in, terminal aesthetic |
| Deployment | Docker + Compose | Reproducible builds, isolated environment, one-command startup |

The deliberate choice to use the standard library for most things — `net/http`, `database/sql`, `context` — is significant. By avoiding ORMs and heavy frameworks, the project maintains full control over connection lifecycles, query execution, and error handling. Chi is the only external HTTP dependency, chosen because it adds routing and middleware without hiding the standard library's interfaces.

## Current Capabilities

As of the latest commit, the system can:

- Register and authenticate users with bcrypt-hashed passwords and JWT tokens.
- Create chat rooms, list them, and join/leave with membership tracking.
- Send and receive real-time messages via WebSockets in joined rooms.
- Persist all messages to PostgreSQL with room-scoped history queries.
- Display join/leave notifications to all room members.
- Serve a terminal-themed single-page application from static files.
- Run entirely inside Docker with automatic database migrations.

## What It Is Not (Yet)

It is important to set expectations. This is not a production chat platform. There is no horizontal scaling (the Hub is a single goroutine in one process), no direct messaging between users, no file uploads, no message editing or deletion, and no rate limiting. The [[03 — WebSocket Real-Time Architecture|WebSocket architecture]] details these gaps honestly.

## Where to Go Next

- To understand the high-level design, read [[02 — Architecture and Design Patterns]].
- To dive into the concurrency model, read [[03 — WebSocket Real-Time Architecture]].
- To understand auth and security, read [[04 — Authentication & Security]].
- To see how it all deploys, read [[05 — Deployment, Build & Operations]].
