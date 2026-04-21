# Raft Consensus Algorithm

### Understanding

From my understanding, Raft is basically an algorithm that solves data consistency and fault tolerance between machines through replication, but there's a lot more caveats and intricacies about it. It also handles leader election, fault recovery, and keeping a strictly ordered log so all machines can stay in sync. 

The reason it's considered a consensus algorithm is because any updates to the machine are only committed when a majority of node agrees, with a leader coordinating the process under a leader-follower architecture.

There's a lot more than this (like log replication details, machine inter-communication), but this is the general idea of what Raft does and solves.

### Hashicorp Raft

For the purpose of this project, this is more to learn how to build a distributed key-value database, and utilize existing tools to create one. Because of this I decided to use `hashicorp/raft`, an external package with an already made Raft implementation.

Basically, through the use of `hashicorp/raft`, all the hard stuff is done. The underlying consensus algorithm is complete, the leader election, log replication, and all the things that Raft does best. What I needed to do is basically integrate that to work with the key-value database I have.

### End to End Write Flow (Set/Del)

In a normal key-value database (one that isn't distributed, say, Redis cache), the data flow would be like:

```plaintext
Client -> kvStore.Set()
       -> db.Update()
```

Basically the client directly writes to the database, and there's nothing happening in between `kvStore.Set()` and `db.Update()` (this is just an example, even though this can get complicated through its own ways).

In a distributed key-value database with Raft, the data flow would look like:

```plaintext
Client -> kvStore.Set()
       -> raft.Apply("set")
       -> FSM.Apply(log)
       -> db.Update()
```

So what’s actually happening here? Basically, Raft receives the command to update the key-value store/database and appends a `"set"` command to its log. It then replicates that log to all other nodes, and once a majority of nodes acknowledge it, the log is committed (meaning it’s now safe and agreed upon). The leader updates its commit index and propagates that to followers, allowing them to also mark the entry as committed. At that point, the committed log entry is passed to the FSM on each node, which then applies it to its own database.

So instead of directly writing to the DB, everything goes through Raft first. Raft ensures that all nodes agree on the operation and the order of operations, and only then is the change applied to the database. This guarantees that all machines stay consistent, even in the presence of failures.

### Useful Links

1. [Raft Visualization](https://thesecretlivesofdata.com/raft/) by [benbjohnson](https://github.com/benbjohnson/thesecretlivesofdata)
2. [Raft Consensus Algorithm](https://raft.github.io/raft.pdf) by Diego Ongaro and John Ousterhout
