---
title: "02 — Architecture and Design Patterns"
date: 2026-05-07
tags: [go, architecture, design-patterns, chi, interfaces, dependency-injection]
---

# 02 — Architecture and Design Patterns

One of the most valuable aspects of `go-chat` is its disciplined use of Go idioms and structural patterns. Despite being a focused project, it follows a clean, layered architecture that would not look out of place in a much larger codebase. This note explores the structural decisions that make the system extensible, testable, and easy to reason about.

## The Three Layers

The system is divided into three primary layers, each with a single, well-defined responsibility:

1.  **HTTP Layer** (`cmd/api/`): Handles routing, middleware, request parsing, and response writing. It knows about HTTP status codes, JSON, and WebSocket upgrades. It knows nothing about SQL.
2.  **Store Layer** (`internal/store/`): Handles database access. It knows about SQL queries, transactions, and PostgreSQL-specific types. It knows nothing about HTTP.
3.  **Infrastructure Layer** (`internal/db/`, `internal/auth/`, `internal/websocket/`): Provides cross-cutting concerns — database connections, password hashing, JWT validation, and real-time message routing.

This separation is crucial. You could swap PostgreSQL for MySQL by reimplementing the store interfaces. You could replace Chi with Gin by rewriting the router setup. The core logic remains untouched.

## Interface-Driven Design

Go's implicit interfaces are used heavily to enforce decoupling.

### The Storage Interface

Defined in `internal/store/storage.go`, the `Storage` struct aggregates all data access contracts:

```go
type Storage struct {
    Users interface {
        Create(context.Context, *User) error
        GetByEmail(context.Context, string) (*User, error)
        GetByID(context.Context, int64) (*User, error)
    }

    Rooms interface {
        Create(context.Context, *Room) error
        GetByID(context.Context, int64) (*Room, error)
        List(context.Context) ([]*Room, error)
        // ...
    }

    Messages interface {
        Create(context.Context, *Message) error
        GetRoomMessages(context.Context, int64, int) ([]*Message, error)
    }

    RoomMembers interface {
        Join(context.Context, int64, int64) error
        Leave(context.Context, int64, int64) error
        IsUserInRoom(context.Context, int64, int64) (bool, error)
    }
}
```

Each field is an interface, not a concrete type. The `application` struct holds a `Storage`, not a `PostgresStorage`. This means handlers are completely decoupled from PostgreSQL. In tests, you can inject a mock `Storage` that returns hardcoded data, eliminating the need for a test database.

### The Application Struct

```go
type application struct {
    config config
    store  store.Storage
    hub    *websocket.Hub
}
```

This is dependency injection in its simplest form. The `application` struct holds everything it needs. Handlers are methods on `application`, so they access `app.store` and `app.hub` directly. There are no package-level variables, no singletons, no hidden globals. This makes the system predictable and trivial to instantiate in tests.

## Why Chi (Not Gin or Echo)

Go has many HTTP routers. Chi was chosen deliberately:

- **Standard library compatibility.** Chi's handlers use `http.HandlerFunc` and `http.Handler`. No custom context types, no wrapping. If you know `net/http`, you know Chi.
- **Middleware composability.** Chi's middleware is just `func(http.Handler) http.Handler`. You can write your own, use Chi's built-ins, or mix standard library middleware. No vendor lock-in.
- **`context.Context` native.** Chi passes `*http.Request` through the standard `context` system. This means timeouts, cancellation, and request-scoped values (like the authenticated user ID) flow naturally through every layer.
- **Lightweight.** Chi is a router and middleware chain. It does not come with rendering, validation, or ORM. You add what you need.

Gin and Echo are excellent frameworks, but they hide the standard library behind their own abstractions. For a project whose goal is to demonstrate understanding of Go's foundational patterns, that opacity is a liability.

## The Middleware Chain

The router mounts a deliberate stack of middleware:

```go
r.Use(middleware.RequestID)      // Unique ID per request for tracing
r.Use(middleware.RealIP)         // Extract real client IP behind proxies
r.Use(middleware.Logger)         // Structured request logging
r.Use(middleware.Recoverer)      // Catch panics, return 500, don't crash
r.Use(middleware.Timeout(60 * time.Second)) // Hard request timeout
```

Then, for protected routes:

```go
r.Group(func(r chi.Router) {
    r.Use(app.AuthMiddleware)    // JWT validation
    // ... protected handlers
})
```

This is defense in depth. `Recoverer` prevents a single handler panic from crashing the server. `Timeout` prevents a slow database query from holding a connection open forever. `AuthMiddleware` runs after the generic stack, so even unauthenticated requests get logged and traced.

## Context Propagation

`context.Context` is the backbone of the system. It flows from the HTTP request through every layer:

```go
// Middleware adds userID to context
ctx := context.WithValue(r.Context(), userIDKey, userID)
next.ServeHTTP(w, r.WithContext(ctx))

// Handler extracts it
userID, _ := GetUserIDFromContext(r.Context())

// Store uses it for timeouts
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel
err := app.store.Messages.Create(ctx, msg)
```

This pattern solves three problems at once:

1.  **Cancellation.** If the client disconnects, `r.Context()` is cancelled, and the database query aborts instead of running to completion.
2.  **Timeouts.** Every store operation receives a context with a deadline. Slow queries fail fast.
3.  **Request tracing.** The `RequestID` middleware adds a unique ID to the context. If you log at every layer, you can trace a single request end-to-end.

## The Options Struct Pattern

Every major component is configured via an options struct:

```go
type config struct {
    addr string
    db   dbConfig
    auth authConfig
}

type dbConfig struct {
    addr         string
    maxOpenConns int
    maxIdleConns int
    maxIdleTime  string
}
```

This is a clean alternative to long constructor parameter lists. It allows for optional configuration, sensible defaults (via `env.GetString` and `env.GetInt`), and forward-compatible APIs. Adding a new config field never changes a function signature.

## Testability

These patterns aren't academic. They solve real problems:

- **You can test handlers without a database.** Mock `store.Storage` and pass it to `application`.
- **You can test the Hub without HTTP.** Create a `Hub` with a mock store and feed it messages directly through channels.
- **You can test auth without a server.** Call `auth.GenerateToken` and `auth.ValidateToken` directly.

This is the difference between a demo and a maintainable system.

## Interview Hook

**Q: "Why interfaces instead of concrete structs?"**

A: Interfaces define behavior, not implementation. In Go, they're implicit — you don't declare that a type satisfies an interface, you just implement the methods. This means I can write a `MockMessageStore` for tests that satisfies the `Messages` interface without importing the real PostgreSQL code. It also means I could swap PostgreSQL for SQLite in tests with a one-line change.

**Q: "Why pass `context.Context` everywhere?"**

A: It's Go's standard mechanism for cancellation, timeouts, and request-scoped values. If a user closes their browser mid-request, `r.Context()` is cancelled, and the database query aborts. Without context, a slow query would hold a connection open until it finished, starving the pool. It's also the only clean way to pass the authenticated user ID from middleware to handlers without global state.

**Q: "What would you change if this grew to 50 handlers?"**

A: I'd split the `application` struct into smaller service structs — `AuthService`, `RoomService`, `MessageService` — each holding only the store interfaces they need. Handlers would depend on services, not the full `Storage` struct. This prevents every handler from having access to every database table, which improves both readability and security.

## Related Notes

- [[01 - Project Overview]]: Goals, capabilities, and tech stack.
- [[03 — WebSocket Real-Time Architecture]]: How the Hub consumes the `Storage` interface for message persistence.
- [[04 - Authentication & Security]]: How `AuthMiddleware` injects user ID into `context.Context`.
- [[05 - Deployment, Build & Operations]]: How the `config` struct is populated from environment variables.
