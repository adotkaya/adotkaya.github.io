---
title: "04 — Authentication & Security"
date: 2026-05-07
tags: [go, jwt, bcrypt, security, authentication, middleware]
---

# 04 — Authentication & Security

Authentication is the gatekeeper. Without it, anyone can read any room's history. Without it, anyone can broadcast messages as anyone else. This note explores how `go-chat` handles identity, from password hashing to JWT validation to the WebSocket upgrade handshake.

## Password Hashing with bcrypt

Passwords are never stored. Only their hashes survive.

```go
func HashPassword(password string) (string, error) {
    hashedBytes, err := bcrypt.GenerateFromPassword(
        []byte(password),
        bcrypt.DefaultCost, // 10 = 2^10 iterations
    )
    return string(hashedBytes), err
}
```

### Why bcrypt

bcrypt is an adaptive hashing function. Its `cost` parameter controls how many iterations it performs. `DefaultCost` is 10, which means 1,024 rounds of the Blowfish cipher. This is intentionally slow — a modern CPU can hash a password in ~100ms. An attacker with a GPU farm can try billions of SHA-256 hashes per second, but only thousands of bcrypt hashes. The asymmetry is the defense.

### Why Not SHA-256 or MD5

SHA-256 is fast. That is its purpose. For passwords, fast is dangerous. A leaked database of SHA-256 password hashes can be cracked offline in hours using rainbow tables. bcrypt salts each password with a random 22-character string, so identical passwords produce different hashes, and rainbow tables are useless.

## JWT Token Design

JSON Web Tokens are stateless. The server does not store session data. It signs a payload and sends it to the client. The client sends it back with every request. The server validates the signature and trusts the claims.

```go
type Claims struct {
    UserID int64 `json:"user_id"`
    jwt.RegisteredClaims
}

func GenerateToken(userID int64, secret string) (string, error) {
    claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "go-chat",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}
```

### Token Structure

A JWT has three parts, base64-encoded and joined by dots:

1.  **Header:** `{"alg":"HS256","typ":"JWT"}`
2.  **Payload:** `{"user_id":42,"exp":...,"iss":"go-chat"}`
3.  **Signature:** `HMAC-SHA256(header + "." + payload, secret)`

The signature is what makes the token tamper-evident. If a user changes their `user_id` to `1` (the admin), the signature no longer matches, and validation fails.

### Why 24 Hours

Shorter expiration limits the window of damage if a token is stolen. A leaked 24-hour token is usable for a day; a leaked 30-day token is usable for a month. The trade-off is user experience — shorter tokens require more frequent logins. In a production system, you'd pair short-lived access tokens with long-lived refresh tokens.

### Why HMAC-SHA256

HMAC (Hash-based Message Authentication Code) is a symmetric algorithm. The same secret signs and verifies. This is simple and fast, but it requires all servers to share the same secret. For distributed systems, asymmetric algorithms (RSA, ECDSA) are preferable because the verifying servers don't need the private key.

## The Auth Middleware

```go
func (app *application) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            writeError(w, http.StatusUnauthorized, "missing authorization header")
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            writeError(w, http.StatusUnauthorized, "invalid header format")
            return
        }

        userID, err := auth.ValidateToken(parts[1], app.config.auth.jwtSecret)
        if err != nil {
            writeError(w, http.StatusUnauthorized, "invalid token")
            return
        }

        ctx := context.WithValue(r.Context(), userIDKey, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Design Decisions

**Bearer scheme.** The `Authorization: Bearer <token>` header is the OAuth 2.0 standard. It's unambiguous and universally supported by HTTP clients.

**Context injection.** The middleware does not attach the user to a global map. It injects the `userID` into `context.Context`, which is request-scoped. When the request ends, the context is garbage-collected. No memory leak, no stale sessions.

**Custom context key type.** `type contextKey string` prevents collisions. If two packages both use `"userID"` as a context key, one overwrites the other. A custom unexported type guarantees uniqueness.

## WebSocket Authentication

WebSocket connections begin as HTTP requests. The client sends an HTTP request to `/v1/rooms/{roomID}/ws` with an `Upgrade: websocket` header. The server validates the request, upgrades the connection, and the persistent socket is born.

But the WebSocket upgrade request must carry the JWT somehow. The project uses the **query parameter** approach:

```
ws://localhost:8080/v1/rooms/5/ws?token=eyJhbGci...
```

### Why Query Parameter

The WebSocket JavaScript API (`new WebSocket(url)`) does not allow custom headers. You cannot send `Authorization: Bearer ...` during the upgrade. The alternatives are:

1.  **Query parameter:** Simple, works everywhere. Risk: token appears in server logs and browser history.
2.  **Cookie:** Sent automatically by the browser. But cookies require CSRF protection and don't work well for SPAs with cross-origin setups.
3.  **Subprotocol:** Overloads the `Sec-WebSocket-Protocol` header. Non-standard and confusing.

The query parameter is the pragmatic choice for this project. In production, you'd use short-lived tokens specifically for WebSocket connections, or switch to cookies with SameSite and Secure flags.

### Room Membership Check

Before upgrading, the handler verifies the user is a member of the room:

```go
inRoom, _ := app.store.RoomMembers.IsUserInRoom(r.Context(), roomID, userID)
if !inRoom {
    http.Error(w, "not a member", http.StatusForbidden)
    return
}
```

This prevents users from connecting to rooms they haven't joined. The check happens at the HTTP layer, before the expensive WebSocket upgrade.

## SQL Injection Prevention

All database queries use parameterized statements:

```go
query := `INSERT INTO messages (room_id, user_id, content) VALUES ($1, $2, $3)`
_, err := db.ExecContext(ctx, query, msg.RoomID, msg.UserID, msg.Content)
```

The `$1, $2, $3` placeholders are bound by the driver. User input is never concatenated into SQL strings. This is the only acceptable pattern for database access. String concatenation, even with escaping, is a bug waiting to happen.

## What This Is Not (Yet)

- **No refresh tokens.** Tokens expire in 24 hours and require re-login. Refresh tokens would allow silent renewal.
- **No rate limiting.** An attacker can attempt unlimited logins. `bcrypt` slows brute-force, but account lockout or CAPTCHA would help.
- **No HTTPS enforcement.** The Docker setup serves HTTP. In production, TLS termination is mandatory.
- **No audit logging.** Failed logins, token validation errors, and suspicious room joins are logged to stdout but not persisted for forensics.

## Interview Hook

**Q: "Why JWT instead of session cookies?"**

A: JWTs are stateless. The server doesn't store session data, which means any server instance can validate a token without a shared session store. This scales horizontally — Server 1 issues a token, Server 2 validates it, no Redis required. The trade-off is revocation: you cannot invalidate a JWT before it expires without a blacklist. For chat, where sessions are short and re-login is acceptable, stateless is the right call.

**Q: "How would you handle token theft?"**

A: Short expiration is the first line of defense — a stolen token is only usable for 24 hours. Second, I'd add refresh tokens with rotation: each refresh issues a new access token *and* a new refresh token, invalidating the old one. If an attacker steals a refresh token and the legitimate user refreshes first, the attacker's token becomes invalid. Finally, binding tokens to IP or device fingerprint would detect anomalies.

**Q: "Why is bcrypt cost 10? Would you increase it?"**

A: 10 is the Go default and a reasonable balance — ~100ms per hash on modern hardware. I'd monitor login latency and increase it to 12 or 14 as hardware improves. The OWASP recommendation is to tune cost so hashing takes at least 250ms. The key is that it's adaptive: existing hashes don't break when you increase the cost, because the cost is stored as part of the hash string.

## Related Notes

- [[01 - Project Overview]]: What the system does and why it exists.
- [[02 - Architecture and Design Patterns]]: How `AuthMiddleware` fits into the middleware chain and context propagation.
- [[03 — WebSocket Real-Time Architecture]]: How JWTs are passed during the WebSocket upgrade handshake.
- [[05 - Deployment, Build & Operations]]: How `JWT_SECRET` is injected via environment variables.
