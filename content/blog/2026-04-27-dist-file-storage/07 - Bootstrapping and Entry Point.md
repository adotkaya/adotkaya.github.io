---
title: "07 - Bootstrapping and Entry Point"
tags: [main, bootstrap, entry-point, makefile, dev-harness]
---

# 07 - Bootstrapping and Entry Point

Every system needs an entry point. For `dist_file_storage`, that entry point is `main.go`, accompanied by a simple but effective `Makefile`. Together, they form the developer experience layer of the project—a dev harness that allows you to spin up a miniature P2P network on your local machine in seconds.

This note walks through `main.go`, the build process, and how the pieces fit together to create a running demo.

## The Goal of main.go

`main.go` is not a production deployment script. It is a **development harness**. Its purpose is to prove that the components—[[03 - Content Addressable Storage|Storage]], [[04 - The P2P Network Layer|Transport]], and [[06 - Server Orchestration|Server]]—can actually work together.

It achieves this by:
1.  Creating two `FileServer` instances.
2.  Configuring them to listen on different TCP ports (`:3000` and `:4000`).
3.  Telling Node 2 to bootstrap to Node 1.
4.  Triggering a `StoreData` call on Node 2 to observe the broadcast.

## Creating the Nodes

The code in `main.go` likely looks something like this:

```go
func main() {
    // Node 1
    tcpTransportOpts1 := p2p.TCPTransportOpts{
        ListenAddr:    ":3000",
        HandshakeFunc: p2p.NOPHandshakeFunc,
        Decoder:       p2p.DefaultDecoder{},
    }
    tr1 := p2p.NewTCPTransport(tcpTransportOpts1)
    
    s1 := &FileServer{
        FileServerOpts: FileServerOpts{
            StorageRoot:       "storage_3000",
            PathTransformFunc: CASPathTransformFunc,
            Transport:         tr1,
            ListenAddr:        ":3000",
        },
        store: NewStore(...),
    }
    
    // Node 2 (bootstraps to Node 1)
    tcpTransportOpts2 := p2p.TCPTransportOpts{
        ListenAddr:    ":4000",
        HandshakeFunc: p2p.NOPHandshakeFunc,
        Decoder:       p2p.DefaultDecoder{},
    }
    tr2 := p2p.NewTCPTransport(tcpTransportOpts2)
    
    s2 := &FileServer{
        FileServerOpts: FileServerOpts{
            StorageRoot:       "storage_4000",
            PathTransformFunc: CASPathTransformFunc,
            Transport:         tr2,
            ListenAddr:        ":4000",
            BootstrapNodes:    []string{":3000"},
        },
        store: NewStore(...),
    }
}
```

(Note: Exact constructor names may vary slightly; this is representative of the pattern.)

## The Bootstrap Process

Node 1 starts listening on `:3000`. Node 2 starts listening on `:4000` and then attempts to dial `:3000`. Because the `FileServer` calls `bootstrapNetwork` inside `Start()`, this connection attempt happens automatically.

Once the TCP connection is established:
1.  `TCPTransport` accepts the connection.
2.  `handleConn` calls the `OnPeer` callback.
3.  `FileServer.OnPeer` adds the new peer to its `peers` map.
4.  Node 1 now knows about Node 2, and Node 2 knows about Node 1.

## The StoreData Demonstration

After starting both servers, `main.go` likely calls:

```go
s2.StoreData("myprivatedata", someReader)
```

This triggers the full lifecycle described in [[06 - Server Orchestration]]:
1.  Node 2 writes the data to its local disk at `storage_4000/...`.
2.  Node 2 constructs a `Payload{Key: "myprivatedata", Data: ...}`.
3.  Node 2 encodes the payload with `gob` and broadcasts it to all peers (which includes Node 1).
4.  Node 1 receives the raw bytes in its transport loop and prints them.

## The Makefile

The `Makefile` provides standard commands:

```makefile
build:
    go build -o bin/fs .

run:
    go run .

test:
    go test ./... -v
```

These are simple but essential. `make build` compiles the binary to `bin/fs`. `make run` executes the dev harness. `make test` runs the unit tests.

In a future iteration, the `Makefile` might expand to include:
- `make docker-build`
- `make lint`
- `make proto` (for Protocol Buffers)

## Observing the System

When you run `make run`, you should see console output from both nodes. Node 1 will print a message indicating it received an `RPC` from Node 2. If you check the `storage_3000` and `storage_4000` directories, you should find identical files stored under the same CAS path, confirming that the broadcast worked.

## Limitations of the Harness

It is important to recognize what `main.go` does *not* do:
- It does not parse command-line flags.
- It does not read a configuration file.
- It is hardcoded to two local nodes.
- It does not demonize or run as a background service.

This is perfectly fine for Phase 1. The harness proves the concept. A production entry point would likely use the `flag` or `cobra` package to accept arguments like `--listen`, `--bootstrap`, and `--storage-root`.

## Related Notes

- [[06 - Server Orchestration]]: Understand what `StoreData` and `Start` actually do.
- [[04 - The P2P Network Layer]]: Understand the TCP transport that `main.go` configures.
- [[09 - Current State and Future Roadmap]]: See where the entry point and deployment strategy could evolve.
