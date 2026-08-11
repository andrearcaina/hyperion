# Distributed KV DB Information

### Definition

At its simplest, a distributed key-value store is like a hash map that runs across multiple machines, or nodes. Its data can be copied across nodes through replication, split across nodes through sharding, or both. Coordination methods such as consensus or quorum-based systems are used when the nodes need to agree.

Conceptually, it exposes a minimal interface:

```
put(key, value)
get(key)
delete(key)
```

In more advanced systems like [Redis](https://redis.io/), values are not just opaque blobs. They can also be structured data types, such as lists, sets, and sorted sets, which support their own operations.

### KV Database System Design

Distributed KV stores generally use either of these approaches, or both:

- Sharding (partitioning): splits data across nodes for scalability, meaning the system can handle more data or traffic by adding more machines
- Replication: copies data across nodes for fault tolerance, meaning the data can remain available when some machines fail

This leads to three common architectures:

1. Replicated (HA) stores
    - Full dataset on every node
    - Focus: availability and consistency
    - Example: [etcd](https://etcd.io/)

    Concept example (3 nodes, entries A through E):

    ```plaintext
    Node 1 (leader)   -> {A: 1, B: 2, C: 3, D: 4, E: 5}
    Node 2 (follower) -> {A: 1, B: 2, C: 3, D: 4, E: 5}
    Node 3 (follower) -> {A: 1, B: 2, C: 3, D: 4, E: 5}
    ```

2. Sharded (clustered) stores
    - Data split across nodes
    - Focus: horizontal scalability
    - Example: [Redis Cluster](https://redis.io/docs/latest/operate/oss_and_stack/management/scaling/)

    Concept example (3 nodes, 3 shards, entries A through E):

    ```plaintext
    Node 1 -> Shard 1 {A: 1, B: 2}
    Node 2 -> Shard 2 {C: 3, D: 4}
    Node 3 -> Shard 3 {E: 5}
    ```

3. Sharded and replicated systems
    - Data partitioned and replicated
    - Focus: scalability and fault tolerance
    - Example: [TiKV](https://tikv.org/) or [CockroachDB](https://www.cockroachlabs.com/)

    Concept example (3 nodes, 3 shards, entries A through E):

    ```plaintext
    Node 1
    ├── leader for Shard 1 {A: 1, B: 2}
    ├── has replica of Shard 2 {C: 3, D: 4}
    └── has replica of Shard 3 {E: 5}

    Node 2
    ├── leader for Shard 2 {C: 3, D: 4}
    ├── has replica of Shard 1 {A: 1, B: 2}
    └── has replica of Shard 3 {E: 5}

    Node 3
    ├── leader for Shard 3 {E: 5}
    ├── has replica of Shard 1 {A: 1, B: 2}
    └── has replica of Shard 2 {C: 3, D: 4}
    ```

    In this example, every node stores every shard, but each node leads a different one. This can be configured differently: a system can use fewer replicas or place them on different nodes, depending on its storage and fault-tolerance needs.

Replication and sharding solve different problems. Replication creates extra copies of data, while sharding divides the data. A system can use either one without the other, or combine them.

### Hyperion Project Specifics

This project draws inspiration from [etcd](https://etcd.io/) and follows a replicated, non-sharded design. It uses the [Raft Consensus Algorithm](https://raft.github.io/raft.pdf) to turn the key-value store into a replicated state machine. Each node keeps a copy of the full dataset rather than owning a separate shard.

### Useful Links

1. [Redis Explained](https://architecturenotes.co/p/redis) by Mahdi Yusuf
2. [Cluster Architecture](https://redis.io/technology/redis-enterprise-cluster-architecture/) by Redis
3. [Deep Dive into etcd: A Distributed Key-Value Store](https://medium.com/@extio/deep-dive-into-etcd-a-distributed-key-value-store-a6a7699d3abc) by Extio Technology
4. [Raft Consensus Algorithm](https://raft.github.io/raft.pdf) by Diego Ongaro and John Ousterhout
