# Raft Consensus Algorithm

### Understanding

From my understanding, Raft is basically an algorithm that solves data consistency and fault tolerance between machines through replication, but there are a lot more caveats and intricacies to it. It also handles leader election, fault recovery, and keeping a strictly ordered log so all machines can stay in sync.

The reason it is considered a consensus algorithm is that updates are only committed when a majority of nodes agree, with a leader coordinating the process under a leader-follower architecture. If a majority of nodes cannot communicate, the cluster cannot safely accept new updates until a majority is available again.

There is a lot more to it, such as the details of log replication and communication between machines, but this is the general idea of what Raft does and the problems it solves.

### HashiCorp Raft

For the purpose of this project, this is more to learn how to build a distributed key-value database, and utilize existing tools to create one. Because of this I decided to use `hashicorp/raft`, an external package with an already made Raft implementation.

Basically, through the use of `hashicorp/raft`, most of the hard parts are already done. The underlying consensus algorithm, leader election, log replication, and other core parts of Raft are already implemented. What I needed to do was integrate it with the key-value database I have.

### End-to-End Write Flow (Set/Del)

In a normal key-value database that is not distributed, the data flow might look like this:

```plaintext
Client -> kvStore.Set()
       -> db.Update()
```

Basically, the client writes directly to the database, and there is nothing happening between `kvStore.Set()` and `db.Update()` in this simplified example. A real local database can still have more steps internally.

In a distributed key-value database that uses Raft, the data flow might look like this:

```plaintext
Client -> kvStore.Set()
       -> raft.Apply("set")
       -> FSM.Apply(log)
       -> db.Update()
```

So what is actually happening here? Basically, Raft receives the command to update the key-value store and appends a `"set"` command to its log. It then replicates that log entry to the other nodes. Once a majority of nodes acknowledge the entry, it can be committed, meaning the cluster has agreed on it. The leader updates its commit index and shares that progress with the followers, allowing them to commit the entry too. The committed entry is then passed to the FSM on each node, which applies it to that node's own database. A slower or unavailable follower may apply the entry later when it catches up.

So instead of writing directly to the database, the update goes through Raft first. Raft makes sure a majority agrees on the operation and that every node applies operations in the same order. This allows the nodes to stay consistent even when some failures happen, as long as a majority of the cluster is still available.

### Useful Links

1. [Raft Visualization](https://thesecretlivesofdata.com/raft/) by [benbjohnson](https://github.com/benbjohnson/thesecretlivesofdata)
2. [Raft Consensus Algorithm](https://raft.github.io/raft.pdf) by Diego Ongaro and John Ousterhout
