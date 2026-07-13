# hyperion

### Overview

Hyperion is a Raft-based, distributed, replicated key-value store built as a learning project, not a production database. It deliberately has one
Raft group, no sharding, no authentication, and no TLS.

### How it works

```text
                          Raft TCP :9001
HTTP :8080 ─┐          ┌─────────────────┐
            ├─> Store ─> replicated log ─> BadgerDB
gRPC :8081 ─┘          └─────────────────┘
```

- `hyprd` runs the HTTP API, gRPC API, Raft node, and local Badger database.
- `hyprctl` talks to either public API.
- The HashiCorp Raft library handles elections and log replication.
- BoltDB stores Raft's log and metadata. BadgerDB stores the user-visible state.
- Writes are accepted only by the leader and return after the command is
  committed and applied locally.
- Reads are also leader-only.

The Raft port is an internal node-to-node protocol. It is not the gRPC port.

### Run one node

```bash
make build
./bin/hyprd --bootstrap
```

In another terminal, use HTTP (the default):

```bash
./bin/hyprctl set greeting hello
./bin/hyprctl get greeting
./bin/hyprctl get
./bin/hyprctl del greeting
```

Or use gRPC:

```bash
./bin/hyprctl --protocol grpc --addr 127.0.0.1:8081 set greeting hello
./bin/hyprctl --protocol grpc --addr 127.0.0.1:8081 get greeting
```

The server enables standard gRPC health checking and server reflection, so tools
such as `grpcurl` can discover the API:

```bash
grpcurl -plaintext 127.0.0.1:8081 list
grpcurl -plaintext -d '{"key":"greeting"}' \
  127.0.0.1:8081 hyperion.v1.Hyperion/Get
```

### Run a three-node cluster

Use a different data directory automatically selected by each node ID, and a
unique public and Raft port for every process:

```bash
./bin/hyprd --node-id n1 --node-addr 127.0.0.1:9001 \
  --srv-port :8080 --grpc-addr :8081 --bootstrap

./bin/hyprd --node-id n2 --node-addr 127.0.0.1:9002 \
  --srv-port :8082 --grpc-addr :8083

./bin/hyprd --node-id n3 --node-addr 127.0.0.1:9003 \
  --srv-port :8084 --grpc-addr :8085
```

Then ask the leader to add each follower:

```bash
./bin/hyprctl join --node-id n2 --node-addr 127.0.0.1:9002
./bin/hyprctl join --node-id n3 --node-addr 127.0.0.1:9003
```

Public requests sent to a follower fail with a clear "not leader" error.

### APIs

The HTTP API remains available under `/hypr`:

| Method | Path | Operation |
| --- | --- | --- |
| `PUT` | `/hypr/kv/{key}` | set a value from the raw request body |
| `GET` | `/hypr/kv/{key}` | get a value |
| `DELETE` | `/hypr/kv/{key}` | delete a value (idempotent) |
| `GET` | `/hypr/kv/` | list all values |
| `POST` | `/hypr/raft/join` | add a Raft voter |

The protobuf contract is in
[`proto/hyperion.proto`](proto/hyperion.proto). Regenerate Go
bindings with `make generate` after changing it.

### Development

Below is a list of development commands:

```bash
make test
make build
```

### Other Information

Data is stored under `~/.hyperion/data/<node-id>`.

For more background, read [the Raft notes](docs/raft.md) and
[the distributed KV overview](docs/distributed-kv-db.md).

### Roadmap

- [x] Create file structure
- [x] Get working DB and spin HTTP server with `hyprd`
- [x] Implement Raft for consensus between nodes (single node for now)
    - [x] Add support for clustering and replication
- [x] Implement `hyprctl` CLI to interact with running `hyprd` nodes
- [X] Add gRPC API support because why not
- [ ] Add Docker and Kubernetes support instead of just processes
- [x] Add documentations and useful things I learnt (upkeep as much as possible)
