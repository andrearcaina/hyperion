package store

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type Node struct {
	raft      *raft.Raft
	transport *raft.NetworkTransport
	store     io.Closer
	cfg       NodeConfig
	logger    *logger.Logger
}

type NodeConfig struct {
	NodeID       string
	NodeAddr     string
	NodePath     string
	ApplyTimeout time.Duration
}

func NewNode(fsm raft.FSM, logger *logger.Logger, cfg *NodeConfig) (*Node, error) {
	if fsm == nil {
		return nil, errors.New("raft FSM is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	if cfg == nil {
		return nil, errors.New("node config is required")
	}

	if cfg.NodeID == "" || cfg.NodeAddr == "" || cfg.NodePath == "" {
		return nil, errors.New("node ID, Raft address, and data path are required")
	}

	if cfg.ApplyTimeout <= 0 {
		return nil, errors.New("apply timeout must be positive")
	}

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
		_ = boltStore.Close()
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
		_ = boltStore.Close()
		return nil, err
	}

	r, err := raft.NewRaft(raftCfg, fsm, boltStore, boltStore, snapStore, transport)
	if err != nil {
		_ = transport.Close()
		_ = boltStore.Close()
		return nil, err
	}

	return &Node{
		raft:      r,
		transport: transport,
		store:     boltStore,
		cfg:       *cfg,
		logger:    logger,
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
		return n.notLeaderError()
	}
	if nodeID == "" || nodeAddress == "" {
		return errors.New("node ID and Raft address are required")
	}

	configuration := n.raft.GetConfiguration()
	if err := configuration.Error(); err != nil {
		return fmt.Errorf("read Raft configuration: %w", err)
	}

	// for each server in the configuration, remove any stale entries for the same ID or address
	for _, server := range configuration.Configuration().Servers {
		if server.ID == raft.ServerID(nodeID) && server.Address == raft.ServerAddress(nodeAddress) {
			return nil
		}

		if server.ID == raft.ServerID(nodeID) || server.Address == raft.ServerAddress(nodeAddress) {
			if err := n.raft.RemoveServer(server.ID, 0, n.cfg.ApplyTimeout).Error(); err != nil {
				return fmt.Errorf("remove stale Raft member %q: %w", server.ID, err)
			}
		}
	}

	// otherwise, add the voter to the cluster
	if err := n.AddVoter(nodeID, nodeAddress); err != nil {
		if errors.Is(err, raft.ErrNotLeader) || errors.Is(err, raft.ErrLeadershipLost) {
			return n.notLeaderError()
		}
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
		n.cfg.ApplyTimeout,
	).Error()
}

func (n *Node) VerifyLeader() error {
	if err := n.raft.VerifyLeader().Error(); err != nil {
		return n.notLeaderError()
	}

	if err := n.raft.Barrier(n.cfg.ApplyTimeout).Error(); err != nil {
		return fmt.Errorf("wait for applied Raft log: %w", err)
	}

	return nil
}

func (n *Node) notLeaderError() error {
	leaderAddress, leaderID := n.raft.LeaderWithID()
	return &NotLeaderError{
		NodeID:        n.cfg.NodeID,
		LeaderID:      string(leaderID),
		LeaderAddress: string(leaderAddress),
	}
}

// wrappers basically
func (n *Node) Apply(data []byte) raft.ApplyFuture { return n.raft.Apply(data, n.cfg.ApplyTimeout) }
func (n *Node) GetNodeID() string                  { return n.cfg.NodeID }
func (n *Node) GetState() raft.RaftState           { return n.raft.State() }
func (n *Node) IsLeader() bool                     { return n.GetState() == raft.Leader }
func (n *Node) Close() error {
	var errs []error
	if err := n.raft.Shutdown().Error(); err != nil {
		errs = append(errs, err)
	}

	if err := n.transport.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := n.store.Close(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
