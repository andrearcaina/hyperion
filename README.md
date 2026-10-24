# Hyperion

Hyperion is a replicated key-value store built to explore distributed systems
and Raft. It is a learning project, not a production database: it has one Raft
group, no sharding, no authentication, and no TLS.

### How it works

Every `hyprd` process is a complete node. It exposes HTTP and gRPC APIs, stores
the user-visible key-value data in BadgerDB, and participates in one Raft group.
The HTTP and gRPC APIs feed into the same store. The leader serves linearizable
reads and sends writes through Raft, which replicates them to a majority before
each node applies them to its local store. BoltDB stores Raft's local state.

```text
                         Raft TCP :9001
HTTP :8080 ─┐          ┌────────────────┐
            ├─> Store ─> replicated log ─> BadgerDB
gRPC :8081 ─┘          └────────────────┘
```

Clients can connect to any node. If a request reaches a follower, the HTTP
handler reverse-proxies it to the current leader, while the gRPC handler
forwards the same RPC. The leader then serves the read or coordinates the
write through Raft.

```text
                         forwarded request
Client ──────> Follower ───────────────────> Leader ──────> Store ──────> Raft
               │                             │
               ├─ HTTP reverse proxy         ├─ HTTP handler
               └─ gRPC forwarding            └─ gRPC handler
```

Data is stored under `~/.hyperion/data/<node-id>`.

For background on the design, see the brief guides in [docs/](docs/):
[distributed key-value stores](docs/distributed-kv-db.md), [Raft](docs/raft.md),
[CAP](docs/cap-theorem.md), and [PACELC](docs/pacelc-theorem.md).

### How to develop locally

Make sure Git is installed, and WSL/Linux is preferered. Hyperion is developed
and tested in a Unix-like environment; its Make targets and shell scripts assume
tools such as `sh`, `make`, and standard Unix commands. Haven't tested on a Windows
environment.

You will also need:

- Go 1.26.7 or a compatible newer version
- Docker and Docker Compose for the local cluster and integration tests
- `kubectl` and [`kind`](https://kind.sigs.k8s.io/) for the Kubernetes workflow

Clone the repository and enter its directory:

```bash
git clone https://github.com/andrearcaina/hyperion.git
cd hyperion
```

Before running a database, validate the development environment:

```bash
make check
```

This runs formatting and imports maintenance, code generation, tests, vet,
and the builds. Other useful development commands are:

Other useful development commands are:

```bash
make generate # regenerate protobuf bindings
make clean    # remove local binaries
```

### Run with Docker Compose

Start a five-node Hyperion database cluster:

```bash
make docker-run
```

Compose gives each node its own container and data volume, bootstraps node 1,
and joins nodes 2 through 5. The nodes publish HTTP on ports `8080`, `8082`,
`8084`, `8086`, and `8088`; gRPC on `8081`, `8083`, `8085`, `8087`, and `8089`;
and Raft on `9001` through `9005`.

```bash
docker compose -f deployments/docker/compose.yml exec node-1 \
    hyprctl set greeting hello
docker compose -f deployments/docker/compose.yml exec node-1 \
    hyprctl get greeting
```

The image includes both `hyprd` and `hyprctl`. You can also run `make install`
and use the local `hyprctl` installed to access the containers/nodes.

Use `make docker-config` to validate the Compose configuration, and
`make docker-status` or `make docker-logs` to inspect the cluster. Run
`make docker-stop` to preserve its data or `make docker-clean` to delete it.

### Run with Kubernetes (`kind`)

For a local Kubernetes cluster, install Docker, `kubectl`, and
[`kind`](https://kind.sigs.k8s.io/), then run:

```bash
make kube-start
```

This creates a disposable Kubernetes kind cluster, builds and loads
the local image, deploys a five-node Hyperion Raft database as a five-pod
StatefulSet, and bootstraps the Raft group. For interactive access, run
`make kube-forward` in one terminal, then use the Hyperion CLI from another:

```bash
hyprctl --addr http://127.0.0.1:8080 set greeting hello
hyprctl --addr http://127.0.0.1:8080 get greeting
```

Inspect it with `make kube-status` or `make kube-logs pod=hyperion-0`. Remove the
cluster with `make kube-stop`.

The kind deployment uses `emptyDir` storage because the cluster is intended
for integration tests. It is not a production persistence configuration.

### Automated integration tests

Run the Go integration suite with Docker available:

```bash
make test-integration
```

The suite includes a five-node Testcontainers cluster. It performs cross-node
read/write verification, runs the chaos scenarios, and cleans up the temporary
containers and network automatically.

#### Manual chaos testing

With the five-node Hyperion database running (either through Docker or Kubernetes), you can run a chaos test scenario:

```bash
make test-chaos scenarios=<scenario> cluster=<docker|k8s>
```

The implemented scenarios are `network-partition`, `sigkill`, `concurrent-writes`, `leader-churn`, and `all`.
`network-partition` simulates a network partition between the leader and the followers.
`sigkill` sends `SIGKILL/kill -9` to a node container, verifies that the remaining majority can
still commit writes, restarts the killed node, and checks that it catches up.
`concurrent-writes` runs 12 concurrent clients writing 10 keys each, and then verifies every acknowledge write from another node.
`leader-churn` deliberately transfers leadership across the five nodes and back to node 1, while verifying that writes are still committed.
`all` runs all four scenarios in sequence.

If no scenario is passed (i.e. `make test-chaos`), `all` will be run against
Docker by default. Start the five-node Docker cluster first with
`make docker-run`. Chaos remains a manual experiment runner; automated
integration lifecycle belongs under Testcontainers. The Kubernetes partition
scenario requires a network plugin that enforces `NetworkPolicy`.

### Run on local environment

Build and start one node:

```bash
make install
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

#### Run a local three-node Hyperion database cluster

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

| Method   | Path                     | Operation                             |
| -------- | ------------------------ | ------------------------------------- |
| `PUT`    | `/hypr/kv/{key}`         | set a value from the raw request body |
| `GET`    | `/hypr/kv/{key}`         | get a value                           |
| `DELETE` | `/hypr/kv/{key}`         | delete a value (idempotent)           |
| `GET`    | `/hypr/kv/`              | list all values                       |
| `POST`   | `/hypr/raft/join`        | add a Raft voter                      |
| `POST`   | `/hypr/raft/move-leader` | move leadership to a voter            |

`PUT /hypr/kv/{key}` accepts a raw byte request body (up to 4 MiB). Successful
HTTP `PUT`, `GET`, and list responses base64-encode values so arbitrary binary
data is preserved; for example, `{"key":"greeting","value":"aGVsbG8="}`.

The gRPC contract is in [`proto/hyperion/v1/hyperion.proto`](proto/hyperion/v1/hyperion.proto). The
server supports gRPC health checking and reflection.

### Roadmap

- [x] Create file structure
- [x] Get working DB and spin HTTP server with `hyprd`
- [x] Implement Raft for consensus between nodes
    - [x] Add support for clustering and replication
    - [x] Support leader-only linearizable reads and writes
    - [x] Allow requests through any node by forwarding them to the leader
- [x] Implement `hyprctl` CLI to interact with running `hyprd` nodes
- [x] Add gRPC API support
- [x] Add Docker support
- [x] Add Kubernetes support (using kind)
- [x] Build a chaos test harness for Docker
    - [x] Network partitions
    - [x] `SIGKILL/kill -9` a random node
    - [x] Concurrent multi-client writes
    - [x] Leader churn
- [x] Add the chaos test harness for Kubernetes
- [x] Expand the Hyperion database setup to five Raft nodes for Docker and Kubernetes
- [ ] Build and deploy a browser dashboard for visualizing Hyperion and Raft experiments
- [x] Add proper integration tests for entire cluster
- [x] Extend Testcontainers integration coverage to a five-node Raft cluster
- [ ] Automate Kubernetes cluster startup for integration tests
- [x] Add documentations and useful things I learnt (upkeep as much as possible)
