# Hyperion

Hyperion is a replicated key-value store built to explore distributed systems
and Raft. It is a learning project, not a production database: it has one Raft
group, no sharding, no authentication, and no TLS.

### How it works

Every `hyprd` process is a complete node that participates in Raft consensus and stores data in BadgerDB, exposing HTTP and gRPC APIs.
Clients use `hyprctl` to send requests through either API, and both feed into the same store.
Raft elects one node as the leader and replicates each write through a majority of the cluster before applying it to BadgerDB.
BoltDB keeps the Raft state, while BadgerDB holds the user-visible key-value data.

```text
                         Raft TCP :9001
HTTP :8080 ─┐          ┌────────────────┐
            ├─> Store ─> replicated log ─> BadgerDB
gRPC :8081 ─┘          └────────────────┘
```

Clients (Users) can connect to any node. If a request reaches a follower, its HTTP
handler acts as a reverse proxy to the current leader, while its gRPC handler
calls the same RPC on the leader. The leader then serves the linearizable read
or coordinates the write through Raft.

```text
                         forwarded request
Client ──────> Follower ───────────────────> Leader ──────> Store ──────> Raft
               │                            │
               ├─ HTTP reverse proxy        ├─ HTTP handler
               └─ gRPC forwarding           └─ gRPC handler
```

Data is stored under `~/.hyperion/data/<node-id>`.

For more background information, check out [docs/](docs/) for a brief overview of Raft, Distributed KV stores, and CAP theorem.

### Run with Docker Compose

Start a three-node cluster:

```bash
make docker-run
```

Compose gives each node its own container and data volume, bootstraps node 1,
and joins nodes 2 and 3. The nodes publish HTTP on ports `8080`, `8082`, and
`8084`; gRPC on `8081`, `8083`, and `8085`; and Raft on `9001`, `9002`, and `9003`.

```bash
docker compose exec node-1 hyprctl set greeting hello
docker compose exec node-1 hyprctl get greeting
```

The image includes both `hyprd` and `hyprctl`. You can also run `make build`
and use the local `hyprctl` in the `bin/` directory against the Docker cluster.

Use `make docker-config` to validate the Compose configuration, and
`make docker-status` or `make docker-logs` to inspect the cluster. Run
`make docker-stop` to preserve its data or `make docker-clean` to delete it.

#### Chaos testing

With the three-node Compose cluster running, you can run a chaos test scenario:

```bash
make test-chaos scenario=<test-scenario>
```

The implemented scenarios are `network-partition`, `sigkill`, and `all`.
`network-partition` simulates a network partition between the leader and the followers.
`sigkill` sends `SIGKILL/kill -9` to a node container, verifies that the remaining majority can
still commit writes, restarts the killed node, and checks that it catches up.

### Run Locally

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
hyprctl join --node-id n2 --node-addr 127.0.0.1:9002 \
  --http-addr 127.0.0.1:8082 --grpc-addr 127.0.0.1:8083
hyprctl join --node-id n3 --node-addr 127.0.0.1:9003 \
  --http-addr 127.0.0.1:8084 --grpc-addr 127.0.0.1:8085
```

The advertised client addresses are optional when every node uses the same
internal HTTP and gRPC ports (as in Docker Compose). Specify them when several
nodes share a host and therefore listen on different ports.

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

### Roadmap

- [x] Create file structure
- [x] Get working DB and spin HTTP server with `hyprd`
- [x] Implement Raft for consensus between nodes (single node for now)
    - [x] Add support for clustering and replication
    - [x] Support leader-only linearizable reads and writes
    - [x] Allow requests through any node by forwarding them to the leader
- [x] Implement `hyprctl` CLI to interact with running `hyprd` nodes
- [x] Add gRPC API support
- [x] Add Docker support
- [ ] Add Kubernetes support
- [ ] Build a chaos test harness for Docker and Kubernetes
    - [x] Network partitions
    - [x] `SIGKILL/kill -9` a random node
    - [ ] Leader churn
    - [ ] Concurrent multi-client writes
- [ ] Add proper integration tests for entire cluster
- [x] Add documentations and useful things I learnt (upkeep as much as possible)
