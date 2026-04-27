---
title: "03 - Content Addressable Storage"
tags: [storage, cas, sha-1, filesystem, go]
---

# 03 - Content Addressable Storage

At the heart of `dist_file_storage` lies its storage layer: `storage.go`. This is the system's foundation, responsible for the most critical task—persisting data to the local filesystem. What makes this layer fascinating is its adherence to **Content Addressable Storage (CAS)** principles. In a CAS system, the address (or path) of an object is derived directly from its contents, typically via a cryptographic hash.

This note provides a comprehensive breakdown of how the project implements CAS on local disk, the design decisions behind its path transformation strategy, and the lifecycle of a file within the store.

## The Store Struct

The `Store` struct is a minimal but effective abstraction over a directory on disk:

```go
type Store struct {
    Root              string
    PathTransformFunc PathTransformFunc
    mu                sync.RWMutex
}
```

It has three fields:
- `Root`: The base directory where all files are saved.
- `PathTransformFunc`: A function that converts a key into a relative file path.
- `mu`: A `sync.RWMutex` to protect concurrent access.

Notice the lack of complexity. There is no database, no B-tree, no complex indexing. Just a root folder and a function. This simplicity is a feature, not a bug. It makes the system predictable and easy to reason about.

## Path Transformation: The Key to CAS

The project defines a `PathTransformFunc` type:

```go
type PathTransformFunc func(string) string
```

This function takes a key (e.g., a filename or an identifier) and returns the path where the file should be stored, relative to `Root`.

### DefaultPathTransformFunc

The simplest possible implementation is the identity function:

```go
func DefaultPathTransformFunc(key string) string {
    return key
}
```

If you store a file with the key `"myphoto.jpg"`, it is written directly to `Root/myphoto.jpg`. This is useful for simple use cases but has a major drawback: it doesn't scale. If keys are user-provided filenames, you run into filesystem limitations (max files per directory, long name limits, collisions).

### CASPathTransformFunc

This is the star of the show. It implements true content-addressable storage:

```go
func CASPathTransformFunc(key string) string {
    hash := sha1.Sum([]byte(key))
    hashStr := fmt.Sprintf("%x", hash)
    
    blocksize := 5
    slices := len(hashStr) / blocksize
    
    paths := make([]string, slices)
    for i := 0; i < slices; i++ {
        from, to := i*blocksize, (i*blocksize)+blocksize
        paths[i] = hashStr[from:to]
    }
    return strings.Join(paths, "/")
}
```

Let's break down what happens when you store a file with the key `"myprivatedata"`:

1.  The key is hashed using **SHA-1**, producing a 40-character hexadecimal string (e.g., `71056ad8aa...`).
2.  This string is split into segments of **5 characters** each.
3.  The segments are joined with `/` to form a path.

For example, a hash might transform into:
```
71056/ad8aa/bf21/...
```

This approach is brilliant for several reasons:

- **Distribution**: By splitting the hash into a tree-like directory structure, it avoids dumping thousands of files into a single folder. Most filesystems degrade in performance when a directory contains too many entries.
- **Addressability**: The path is deterministic. If you know the key, you can always compute the exact path on disk without a lookup table.
- **Integrity**: If the content changes, the key (and thus the path) changes. This naturally prevents accidental overwrites and makes versioning implicit.
- **Flat Keyspace**: It allows the system to handle arbitrary keys (even very long ones) because the final filename is always a fixed-length hash.

## The Store Lifecycle

The `Store` provides a complete lifecycle for file management:

### Write

`Write(key string, r io.Reader) (int64, error)`:
1.  Computes the path using `PathTransformFunc(key)`.
2.  Creates the necessary directories using `os.MkdirAll`.
3.  Opens the file with `os.Create`.
4.  Copies the data from the `io.Reader` into the file using `io.Copy`.

It returns the number of bytes written. Because it takes an `io.Reader`, it is extremely flexible—you can write from a network connection, a byte buffer, or a local file without changing the method signature.

### Read

`Read(key string) (io.Reader, error)`:
1.  Computes the path.
2.  Opens the file with `os.Open`.
3.  Returns the `*os.File` (which satisfies `io.Reader`).

This is a clean, idiomatic Go pattern.

### Has

`Has(key string) bool`:
Checks if the file exists using `os.Stat`. This is useful for deduplication or cache checks.

### Delete

`Delete(key string) error`:
Removes the file at the computed path.

### Clear

`Clear() error`:
A convenience method that wipes the entire `Root` directory. This is incredibly useful for testing and development.

## Concurrency and Safety

The `Store` uses a `sync.RWMutex`. While the current implementation is relatively simple, the mutex is there to protect against race conditions if multiple goroutines attempt to write to the same key simultaneously or if a read happens during a write. In a future iteration, this might be refined to per-key locking for higher concurrency.

## Connection to the Network

The `Store` is intentionally isolated. It does not know about TCP, peers, or broadcasting. It is wired into the `FileServer` (discussed in [[06 - Server Orchestration]]), which calls `Store.Write` after receiving data from the network, or before broadcasting data out to the network. This separation of concerns is what makes the architecture so clean.

## Related Notes

- [[02 - Architecture and Design Patterns]]: Understand why `PathTransformFunc` is an injectable strategy.
- [[06 - Server Orchestration]]: See how `FileServer` uses the `Store`.
- [[08 - Testing Strategy]]: Learn how `storage_test.go` validates the CAS logic.
