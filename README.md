# hyperion

a Raft-based distributed reliable key-value database

### Overview

currently just spins up a HTTP REST server that wraps BadgerDB

will do more i promise

### Roadmap

- [x] Create file structure
- [x] Get working DB and spin REST server with `hyprd`
- [x] Implement Raft for consensus between nodes (single node for now)
    - [ ] Add support for clustering and replication
- [ ] Implement `hyprctl` CLI to interact with running `hyprd` nodes
- [ ] Add gRPC API support because why not
- [ ] Add documentations and useful things I learnt
