# Brewer's CAP Theorem

### Definition

At it's core, it is a principle for distributed systems that states any distributed system can have at most two of the following properties:

- Consistency (C): Every read receives the most recent successful write, or an error.
- Availability (A): Every request receives a non-error response, without guarantee that it contains the most recent write.
- Partition Tolerance (P): The system continues to operate despite network failures that prevent communication between nodes.

This means that systems are either **CP** (Consistency and Partition Tolerance), **AP** (Availability and Partition Tolerance), or **CA** (Consistency and Availability). However, in practice, CA is not achievable in distributed systems due to the inevitability of network partitions. Thus, most distributed systems must choose between CP and AP. It's mathematically impossible to achieve all three properties simultaneously in a distributed system (check this [reddit link](https://www.reddit.com/r/dataengineering/comments/1kzrhwt/cap_theorem_possible_to_achieve_all_three/), a reputable source).

### Consistency Patterns

- Strong Consistency: all reads see the latest data immediately, making it more reliable, but can be slower.
- Eventual Consistency: data may differ temporarily between servers, but will eventually sync, making it faster but less reliable.
- Weak Consistency: data may be old or inconsistent, with no guarantees of synchronization, making it the fastest but least reliable.

### Availability Patterns

- High Availability: the system is designed to be operational and accessible at all times, often through redundancy and failover mechanisms.
    - 99s Availability: A measure of system uptime, with three to five nines indicating high availability (e.g., 99.999% uptime).
- Low Availability: the system may experience downtime or unavailability, often due to maintenance, failures, or lack of redundancy. (NOT RECOMMENDED)

### CP vs AP Systems

| Model | Prioritizes  | During Partition           | Trade-off                         | Examples                      |
| ----- | ------------ | -------------------------- | --------------------------------- | ----------------------------- |
| CP    | Consistency  | May reject/delay requests  | Sacrifices availability           | Spanner, CockroachDB, etcd    |
| AP    | Availability | Continues serving requests | May return stale/conflicting data | Cassandra, DynamoDB, ScyllaDB |

### Useful Links

1. [System Design Roadmap](https://roadmap.sh/system-design) by Roadmap.sh
2. [What is the CAP theorem?](https://www.ibm.com/think/topics/cap-theorem) by IBM
3. [An Illustrated Proof of the CAP Theorem](https://mwhittaker.github.io/blog/an_illustrated_proof_of_the_cap_theorem/) by Matt Whittaker
