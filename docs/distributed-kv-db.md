# Distributed KV DB Information

### Definition

At its simplest, a distributed key-value store is a hash map where data is distributed across multiple machines (nodes), and kept in sync through replication, sharding, or both, using coordination protocols such as consensus or quorum-based systems.

Conceptually, it exposes a minimal interface:

```
put(key, value)
get(key)
delete(key)
```

In more advanced systems like [Redis](https://redis.io/), values are not just opaque blobs but structured data types (e.g., lists, sets, sorted sets), allowing different operations.

### KV Database System Design

Distributed KV stores generally vary along two types (can be both):

- Sharding (partitioning): splits data across nodes for scalability (ability to handle more data or traffic by adding more machines)
- Replication: duplicates data across nodes for fault tolerance (ability to keep working even when some machines fail)

This leads to three common architectures:

1. Replicated (HA) stores
    - Full dataset on every node
    - Focus: availability and consistency
    - Example: [etcd](https://etcd.io/)
    - Concept Example (3 nodes, keys A, B, C):
        - Node 1 (leader): A B C
        - Node 2 (follower): A B C
        - Node 3 (follower): A B C

2. Sharded (clustered) stores
    - Data split across nodes
    - Focus: horizontal scalability
    - Example: [Redis Cluster](https://redis.io/docs/latest/operate/oss_and_stack/management/scaling/) (unless configured differently)
    - Concept Example (3 nodes, keys A, B, C):
        - Node 1: A
        - Node 2: B
        - Node 3: C

3. Sharded + replicated systems
    - Data partitioned and replicated
    - Focus: scalability + fault tolerance (modern approach)
    - Example: [TiKV](https://tikv.org/) or [CockroachDB](https://www.cockroachlabs.com/)
    - Concept Example (3 nodes, keys A, B, C):
        - Shard 1 (A): Node 1 (leader), Node 2 (replica/follower)
        - Shard 2 (B): Node 2 (leader), Node 3 (replica/follower)
        - Shard 3 (C): Node 3 (leader), Node 1 (replica/follower)

### Hyperion Project Specifics

This project draws inspiration from [etcd](https://etcd.io/) and follows a replicated (non-sharded) design, utilizing [Raft Consensus Algorithm](https://raft.github.io/raft.pdf) to essentially make a key-value store distributed by implementing a replicated state machine.

### Useful Links

1. [Redis Explained](https://architecturenotes.co/p/redis) by Mahdi Yusuf
2. [Cluster Architecture](https://redis.io/technology/redis-enterprise-cluster-architecture/) by Redis
3. [Deep Dive into etcd: A Distributed Key-Value Store](https://medium.com/@extio/deep-dive-into-etcd-a-distributed-key-value-store-a6a7699d3abc) by Extio Technology
4. [Raft Consensus Algorithm](https://raft.github.io/raft.pdf) by Diego Ongaro and John Ousterhout