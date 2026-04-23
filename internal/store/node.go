package store

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type Node struct {
	raft *raft.Raft
	cfg  NodeConfig
	log  *logger.Logger
}

type NodeConfig struct {
	NodeID       string
	ApplyTimeout time.Duration
	DBPath       string
}

func NewNode(fsm raft.FSM, logger *logger.Logger, cfg *NodeConfig) (*Node, error) {
	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.NodeID)

	// using BoltDB as the backend for both the log store and stable store
	// honestly i shouldve used boltdb for data too but i wanted to try out badger for the data store
	boltStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DBPath, "raftbolt.db"))
	if err != nil {
		return nil, err
	}

	snapStore, err := raft.NewFileSnapshotStore(cfg.DBPath, 1, nil)
	if err != nil {
		return nil, err
	}

	// later on use NewTCPTransport for real network communication between nodes
	// this is only for testing (single node raft setup)
	addr, transport := raft.NewInmemTransport(raft.ServerAddress("127.0.0.1:0"))

	r, err := raft.NewRaft(raftCfg, fsm, boltStore, boltStore, snapStore, transport)
	if err != nil {
		return nil, err
	}

	bootstrap := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raftCfg.LocalID,
				Address: addr,
			},
		},
	}

	// bootstraps a single-node Raft cluster (no leader is explicitly assigned)
	if err := r.BootstrapCluster(bootstrap).Error(); err != nil && !errors.Is(err, raft.ErrCantBootstrap) {
		return nil, err
	}

	return &Node{
		raft: r,
		cfg:  *cfg,
		log:  logger,
	}, nil
}

func (n *Node) Apply(data []byte) raft.ApplyFuture {
	return n.raft.Apply(data, n.cfg.ApplyTimeout)
}

func (n *Node) GetID() string {
	return n.cfg.NodeID
}

func (n *Node) GetState() raft.RaftState {
	return n.raft.State()
}

func (n *Node) GetLeader() string {
	return string(n.raft.Leader())
}

func (n *Node) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

func (n *Node) Close() error {
	return n.raft.Shutdown().Error()
}
