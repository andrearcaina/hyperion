# hyperion

a Raft-based distributed reliable key-value database

### Overview

now able to spin up multiple HTTP servers with their separate Raft node now

basically how it works is when you start a `hyprd` node, it will create a HTTP server and a Raft store (node)

the HTTP server will listen for incoming requests (either through `curl` or `hyprctl`) and forward them to Raft

Raft then records those requests as a log entry, sends them to other `hyprd` nodes, agrees on a majority and then applies/commits them to the underlying key-value database (BadgerDB)

for more information on Raft and distributed key value databases check out the [docs](docs/) page

### Roadmap

- [x] Create file structure
- [x] Get working DB and spin HTTP server with `hyprd`
- [x] Implement Raft for consensus between nodes (single node for now)
    - [x] Add support for clustering and replication
- [x] Implement `hyprctl` CLI to interact with running `hyprd` nodes
- [ ] Add gRPC API support because why not
- [ ] Add Docker and Kubernetes support instead of just processes
- [x] Add documentations and useful things I learnt (upkeep as much as possible)
