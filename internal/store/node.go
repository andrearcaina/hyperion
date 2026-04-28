package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type Node struct {
	raft   *raft.Raft
	cfg    NodeConfig
	logger *logger.Logger
}

type NodeConfig struct {
	NodeID       string
	NodeAddr     string
	NodePath     string
	ApplyTimeout time.Duration
}

func NewNode(fsm raft.FSM, logger *logger.Logger, cfg *NodeConfig) (*Node, error) {
	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.NodeID)

	// using BoltDB as the backend for both the log store and stable store
	// honestly i shouldve used boltdb for data too but i wanted to try out badger for the data store
	boltStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.NodePath, "raftbolt.db"))
	if err != nil {
		return nil, err
	}

	snapStore, err := raft.NewFileSnapshotStore(cfg.NodePath, 1, nil)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(
		cfg.NodeAddr,
		nil,
		10,
		10*time.Second,
		os.Stderr,
	)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(raftCfg, fsm, boltStore, boltStore, snapStore, transport)
	if err != nil {
		return nil, err
	}

	return &Node{
		raft:   r,
		cfg:    *cfg,
		logger: logger,
	}, nil
}

func (n *Node) BootstrapCluster() error {
	config := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(n.cfg.NodeID),
				Address: raft.ServerAddress(n.cfg.NodeAddr),
			},
		},
	}

	return n.raft.BootstrapCluster(config).Error()
}

func (n *Node) Join(nodeID, nodeAddress string) error {
	if !n.IsLeader() {
		return fmt.Errorf("node %s is not the leader", n.GetNodeID())
	}

	if err := n.AddVoter(nodeID, nodeAddress); err != nil {
		return err
	}

	n.logger.Info(context.Background(), "successfully added voter to cluster", "node_id", nodeID, "node_address", nodeAddress)
	return nil
}

func (n *Node) AddVoter(nodeID, nodeAddress string) error {
	return n.raft.AddVoter(
		raft.ServerID(nodeID),
		raft.ServerAddress(nodeAddress),
		0,
		10*time.Second,
	).Error()
}

// wrappers basically
func (n *Node) Apply(data []byte) raft.ApplyFuture { return n.raft.Apply(data, n.cfg.ApplyTimeout) }
func (n *Node) GetNodeID() string                  { return n.cfg.NodeID }
func (n *Node) GetState() raft.RaftState           { return n.raft.State() }
func (n *Node) IsLeader() bool                     { return n.GetState() == raft.Leader }
func (n *Node) Close() error                       { return n.raft.Shutdown().Error() }
