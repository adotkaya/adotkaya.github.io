---
title: "08 - Testing Strategy"
tags: [testing, go, unit-tests, testify, tdd]
---

# 08 - Testing Strategy

A distributed system is only as reliable as its tests. While `dist_file_storage` is in early development, it already demonstrates a solid testing philosophy centered on **unit testing** and **behavioral validation**. This note explores the test suite, the tools used, and the areas that will need more rigorous testing as the project matures.

## Test Files Overview

The project contains at least two test files:
- `storage_test.go`
- `p2p/tcp_transport_test.go`

These align with the two most critical layers: storage and networking.

## Testing the Storage Layer (storage_test.go)

The storage layer is the easiest to test because it has no external dependencies (no network, no database). It interacts purely with the local filesystem, making it ideal for table-driven tests.

### CAS Path Transformation Tests

One of the most important things to test is the `CASPathTransformFunc`. If the hashing or path-splitting logic is flawed, the entire content-addressability guarantee is broken.

A test for this might look like:

```go
func TestCASPathTransformFunc(t *testing.T) {
    key := "myprivatedata"
    pathKey := CASPathTransformFunc(key)
    
    // A SHA-1 hash is 40 hex characters.
    // Split into 5-char blocks, that's 8 directories deep.
    parts := strings.Split(pathKey, "/")
    require.Equal(t, 8, len(parts))
    
    // Verify the reconstructed path is deterministic
    require.Equal(t, pathKey, CASPathTransformFunc(key))
}
```

This test validates:
1.  **Determinism**: The same key always produces the same path.
2.  **Structure**: The path has the expected number of segments.
3.  **Length**: Implicitly validates that the SHA-1 hash is being used correctly.

### Store Lifecycle Tests

The test suite likely exercises the full lifecycle of the `Store`:

```go
func TestStore(t *testing.T) {
    s := NewStore(StoreOpts{
        Root:              "test_root",
        PathTransformFunc: CASPathTransformFunc,
    })
    defer teardown(t, s)
    
    key := "my_special_key"
    data := []byte("some jpg bytes")
    
    // Write
    n, err := s.Write(key, bytes.NewReader(data))
    require.NoError(t, err)
    require.Equal(t, int64(len(data)), n)
    
    // Has
    require.True(t, s.Has(key))
    
    // Read
    r, err := s.Read(key)
    require.NoError(t, err)
    
    b, _ := io.ReadAll(r)
    require.Equal(t, data, b)
    
    // Delete
    require.NoError(t, s.Delete(key))
    require.False(t, s.Has(key))
}
```

This is a comprehensive behavioral test. It doesn't just test one method; it tests the contract of the `Store`: that data written can be read back, that `Has` correctly reflects existence, and that `Delete` actually removes the file.

### The teardown Helper

Tests that write to disk must clean up after themselves. A `teardown` function is likely used:

```go
func teardown(t *testing.T, s *Store) {
    if err := s.Clear(); err != nil {
        t.Error(err)
    }
}
```

This uses the `Store.Clear()` method to wipe the test directory, ensuring test isolation.

## Testing the Network Layer (tcp_transport_test.go)

Networking is harder to test than filesystem I/O because it involves concurrency and real system resources (sockets, ports).

The current TCP transport test is likely minimal:

```go
func TestTCPTransport(t *testing.T) {
    opts := TCPTransportOpts{
        ListenAddr:    ":4000",
        HandshakeFunc: NOPHandshakeFunc,
        Decoder:       DefaultDecoder{},
    }
    tr := NewTCPTransport(opts)
    
    require.NoError(t, tr.ListenAndAccept())
    
    // In a real test, you might dial the transport here
    // and verify that an RPC is received on the consume channel.
}
```

This test proves that the transport can bind to a port and start listening without crashing. It is a "smoke test" for the transport's initialization logic.

## Testing Tools

The project uses the standard Go testing toolkit plus `github.com/stretchr/testify v1.11.1`.

### testify

`testify` provides:
- `require.NoError(t, err)`: Fails the test immediately if an error occurs.
- `require.Equal(t, expected, actual)`: Clean, readable assertions.
- `require.True(t, condition)`: For boolean checks.

This library dramatically improves the readability of tests compared to manual `if err != nil { t.Fatal(err) }` blocks.

## What's Missing: The Test Gap

As the project evolves, the test suite will need to expand significantly:

### Integration Tests

Currently, there are no integration tests that spin up multiple nodes and verify end-to-end behavior. An integration test would:
1.  Start Node 1.
2.  Start Node 2 and bootstrap to Node 1.
3.  Call `StoreData` on Node 2.
4.  Assert that Node 1 received the data and wrote it to disk.

This would catch bugs in the broadcasting logic, the peer management, and the transport wiring.

### Concurrency Tests

The `Store` uses a `sync.RWMutex`, and the `FileServer` uses a `sync.Mutex`. These should be stress-tested with many goroutines writing and reading simultaneously. The Go race detector (`go test -race`) should be run regularly.

### Failure Injection

A robust distributed system must be tested under failure. Future tests should:
- Kill a peer mid-broadcast and ensure the server doesn't crash.
- Attempt to dial an offline bootstrap node and verify graceful handling.
- Corrupt a payload and verify that the receiver handles it.

### Fuzzing

Go 1.18+ supports fuzzing. The `PathTransformFunc` and `Decoder` logic are perfect candidates for fuzz tests, which could uncover edge cases with malformed inputs.

## Running the Tests

As defined in the `Makefile`:

```bash
make test
```

This runs `go test ./... -v`, executing all tests in all packages with verbose output.

## Related Notes

- [[03 - Content Addressable Storage]]: The storage logic being tested.
- [[04 - The P2P Network Layer]]: The transport logic being tested.
- [[09 - Current State and Future Roadmap]]: See how testing fits into the project's evolution.
