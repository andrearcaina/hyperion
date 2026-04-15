# hyperion

a Raft-based distributed reliable key-value database

### Overview

now spins up an HTTP server and a single Raft node (for now)

basically how it works is when you start a `hyprd` node, it will create a HTTP server and a Raft store (node)

the HTTP server will listen for incoming requests and forward them to Raft

Raft then records those requests as a log entry, and then applies them to the underlying key-value database (BadgerDB)

the Raft node also handles a lot of other things, like snapshots, log replication (to other nodes), and other stuff that I don't really understand yet but will learn eventually

will do more i promise

### Roadmap

- [x] Create file structure
- [x] Get working DB and spin HTTP server with `hyprd`
- [x] Implement Raft for consensus between nodes (single node for now)
    - [ ] Add support for clustering and replication
- [ ] Implement `hyprctl` CLI to interact with running `hyprd` nodes
- [ ] Add gRPC API support because why not
- [ ] Add documentations and useful things I learnt
