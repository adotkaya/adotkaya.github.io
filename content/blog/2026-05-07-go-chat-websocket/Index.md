---
title: "Building a Real-Time Chat Application in Go"
date: 2026-05-07
tags: [go, chat, websockets, project-deep-dive, backend]
---

# Building a Real-Time Chat Application in Go

Welcome to this comprehensive deep dive into `go-chat`, a real-time chat application built with Go, WebSockets, JWT authentication, and PostgreSQL. This project represents a ground-up exploration of production-grade backend patterns — from goroutine-driven concurrency to interface-driven architecture to Docker-based deployment.

This is not a tutorial, but rather a long-form exploration of the codebase, the design decisions, and the architectural patterns that make the system work. You can navigate through the ideas chronologically or jump to any layer that interests you.

## What You Will Find Here

- **A complete architectural breakdown** of a Go-based real-time system.
- **Deep dives into specific layers**: from HTTP routing to WebSocket concurrency.
- **Design pattern analysis**: why interfaces, channels, and context propagation are used.
- **Security considerations**: bcrypt, JWT, middleware, and SQL injection prevention.
- **Deployment and operations**: multi-stage Docker builds, migrations, and Makefile workflows.
- **An honest assessment** of what works and what would need to change for production.

## Start Here

If you are new to the project, begin with [[01 - Project Overview]] to understand the "what" and the "why". From there, you can follow the logical flow through the architecture or jump to any specific layer that interests you.

## The Structure of This Series

| Note | Focus |
|------|-------|
| [[01 - Project Overview]] | Goals, tech stack, motivation, current capabilities |
| [[02 - Architecture and Design Patterns]] | Layered architecture, interfaces, context, Chi router |
| [[03 — WebSocket Real-Time Architecture]] | Hub pattern, goroutines, fan-out, ping/pong, broadcast |
| [[04 - Authentication & Security]] | bcrypt, JWT, middleware, WebSocket auth, SQL injection |
| [[05 - Deployment, Build & Operations]] | Docker, multi-stage builds, migrations, Makefile |

---

*This series is designed to be read linearly or explored as a graph. Every major concept is linked to its related files so you can trace dependencies across the system.*
