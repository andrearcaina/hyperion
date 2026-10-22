# PACELC Theorem

### Definition

At its core, PACELC expands on the CAP theorem by describing a trade-off that also exists during normal operation:

> If there is a Partition (P), choose between Availability (A) and Consistency (C).
> Else (E), choose between lower Latency (L) and Consistency (C).
>
> — Daniel J. Abadi, adapted from *Consistency Tradeoffs in Modern Distributed Database System Design*

The name comes from the four parts of the trade-off: **P**artition, **A**vailability, **C**onsistency, **E**lse, and **L**atency. [CAP theorem](./cap-theorem.md) describes what happens during a network partition, while PACELC also describes the consistency and latency decision a system makes when the network is healthy.

### PACELC Patterns

The notation describes the choice in each operating condition:

- The first pair, `PA` or `PC`, describes the choice during a **P**artition.
- The second pair, `EL` or `EC`, describes the choice **E**lse, when there is no partition.

For example, `PA/EL` means “choose Availability during a Partition, else choose lower Latency.”

### Latency and Coordination

- Low Latency: requests return quickly, often because the system avoids waiting for remote replicas or quorum acknowledgement. This can provide faster responses but may weaken consistency guarantees.
- Coordination Cost: requests may wait for network round trips, replica acknowledgement, or a quorum before returning. This is not a goal of the system; it is the latency cost of providing stronger consistency.

Latency is not the same as availability. A system can be available but slow, or unavailable because it chooses to wait for consistency. The consistency and availability definitions are described in the [CAP theorem](./cap-theorem.md).

### PACELC Trade-offs

| Model | Prioritizes | During Partition | During Normal Operation | Trade-off |
| ----- | ------------ | ---------------- | ----------------------- | --------- |
| PA/EL | Availability and lower latency | Keeps serving requests | Avoids coordination where possible | Faster responses, weaker consistency guarantees |
| PC/EC | Consistency | May reject or delay requests | Waits for replica or quorum coordination | Stronger guarantees, higher latency |
| PA/EC | Availability and consistency | Keeps serving requests with relaxed guarantees | Coordinates replicas | Available during partitions, consistent otherwise |
| PC/EL | Consistency and lower latency | May reject or delay requests | Avoids coordination where possible | Consistent partition behavior, faster normal responses |

PACELC is not a ranking system. The right choice depends on whether stale data, rejected requests, or additional network latency is more harmful to the application.

### Hyperion Project Specifics

Hyperion uses Raft to replicate its key-value state machine. For replicated writes, it is intended to behave like **PC/EC**:

- During a partition, a node cannot commit a write without a leader and a majority, so the system favors consistency over availability.
- During normal operation, a write is committed through Raft replication before it is acknowledged, so the system favors consistency over the lowest possible latency.

This is why the chaos tests expect the majority to keep accepting writes while an isolated node rejects them, and why the isolated node catches up after the partition is healed.

### Useful Links

1. [Consistency Tradeoffs in Modern Distributed Database System Design](https://www.cs.umd.edu/~abadi/papers/abadi-pacelc.pdf) by Daniel J. Abadi
2. [Brewer's CAP Theorem](./cap-theorem.md)
3. [Raft Consensus Algorithm](./raft.md)
