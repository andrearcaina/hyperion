# hyperion

Hyperion is a replicated key-value store built to explore distributed systems
and Raft. It is a learning project, not a production database: it has one Raft
group, no sharding, no authentication, and no TLS.

### How it works

```text
                         Raft TCP :9001
HTTP :8080 ─┐          ┌────────────────┐
            ├─> Store ─> replicated log ─> BadgerDB
gRPC :8081 ─┘          └────────────────┘
```

- Each `hyprd` process is a complete node: HTTP, gRPC, Raft, and BadgerDB.
- Raft elects one leader and replicates writes through a majority of nodes.
- The leader serves linearizable reads and writes; followers reject client
  requests for now.
- BoltDB stores Raft state, while BadgerDB stores the user-visible key-value
  data.
- `hyprctl` is the client for both the HTTP and gRPC APIs.

Port `9001` is for internal Raft traffic, not gRPC.

### Run with Docker

Start a three-node cluster:

```bash
make docker-run
```

Compose gives each node its own container and data volume, bootstraps node 1,
and joins nodes 2 and 3. The nodes publish HTTP on ports `8080`, `8082`, and
`8084`; gRPC on `8081`, `8083`, and `8085`; and Raft on `9001`–`9003`.

```bash
docker compose exec node-1 hyprctl set greeting hello
docker compose exec node-1 hyprctl get greeting
```

The image includes both `hyprd` and `hyprctl`. You can also run `make build`
and use the local `hyprctl` in the `bin/` directory against the Docker cluster.

Use `make docker-status` or `make docker-logs` to inspect the cluster. Run
`make docker-stop` to preserve its data or `make docker-clean` to delete it.

### Run locally

Build and start one node:

```bash
make build
hyprd --bootstrap
```

Use `hyprctl` from another terminal:

```bash
hyprctl set greeting hello
hyprctl get greeting
hyprctl get
hyprctl del greeting
```

HTTP on `127.0.0.1:8080` is the default. To use gRPC instead, pass
`--protocol grpc --addr 127.0.0.1:8081`.

#### Run a local three-node cluster

Start each node in a separate terminal:

```bash
hyprd --node-id n1 --node-addr 127.0.0.1:9001 \
  --srv-port :8080 --grpc-addr :8081 --bootstrap

hyprd --node-id n2 --node-addr 127.0.0.1:9002 \
  --srv-port :8082 --grpc-addr :8083

hyprd --node-id n3 --node-addr 127.0.0.1:9003 \
  --srv-port :8084 --grpc-addr :8085
```

Then join the followers through node 1:

```bash
hyprctl join --node-id n2 --node-addr 127.0.0.1:9002
hyprctl join --node-id n3 --node-addr 127.0.0.1:9003
```

### Interfaces

The HTTP API is under `/hypr`:

| Method   | Path              | Operation                             |
| -------- | ----------------- | ------------------------------------- |
| `PUT`    | `/hypr/kv/{key}`  | set a value from the raw request body |
| `GET`    | `/hypr/kv/{key}`  | get a value                           |
| `DELETE` | `/hypr/kv/{key}`  | delete a value (idempotent)           |
| `GET`    | `/hypr/kv/`       | list all values                       |
| `POST`   | `/hypr/raft/join` | add a Raft voter                      |

The gRPC contract is in [`proto/hyperion.proto`](proto/hyperion.proto). The
server supports gRPC health checking and reflection.

### Development

```bash
make check    # generate, vet, test, and build
make generate # regenerate protobuf bindings
make clean    # remove local binaries
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
    - [x] Support leader-only linearizable reads and writes
    - [ ] Allow requests through any node by forwarding them to the leader
- [x] Implement `hyprctl` CLI to interact with running `hyprd` nodes
- [x] Add gRPC API support
- [x] Add Docker support
- [ ] Add Kubernetes support
- [ ] Build a chaos test harness for Docker and Kubernetes
    - [ ] Network partitions
    - [ ] Leader churn
    - [ ] `kill -9` crashes
    - [ ] Concurrent multi-client writes
- [x] Add documentations and useful things I learnt (upkeep as much as possible)
