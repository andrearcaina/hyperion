package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/hashicorp/raft"
)

var (
	ErrInvalidKey = errors.New("key must not be empty")
	ErrNotFound   = db.ErrNotFound
)

type NotLeaderError struct {
	NodeID        string
	LeaderID      string
	LeaderAddress string
}

func (e *NotLeaderError) Error() string {
	if e.LeaderID == "" {
		return fmt.Sprintf("node %q is not the Raft leader; no leader is known", e.NodeID)
	}

	return fmt.Sprintf("node %q is not the Raft leader; leader is %q at %s", e.NodeID, e.LeaderID, e.LeaderAddress)
}

type Store struct {
	db     *db.DB
	node   *Node
	logger *logger.Logger
}

func New(db *db.DB, logger *logger.Logger, cfg *NodeConfig) (*Store, error) {
	fsm, err := NewFSM(db)
	if err != nil {
		return nil, err
	}

	node, err := NewNode(fsm, logger, cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		db:     db,
		node:   node,
		logger: logger,
	}, nil
}

func (s *Store) Set(key string, value []byte) error {
	if key == "" {
		return ErrInvalidKey
	}

	return s.applyCommand(writeCommand{
		Op:    commandSet,
		Key:   key,
		Value: value,
	})
}

func (s *Store) Delete(key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	return s.applyCommand(writeCommand{
		Op:  commandDelete,
		Key: key,
	})
}

func (s *Store) Join(nodeID, nodeAddress string) error {
	return s.node.Join(nodeID, nodeAddress)
}

func (s *Store) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	if err := s.node.VerifyLeader(); err != nil {
		return nil, err
	}

	return s.db.Get([]byte(key))
}

func (s *Store) ForEach(fn func(key, value []byte) error) error {
	if err := s.node.VerifyLeader(); err != nil {
		return err
	}

	return s.db.ForEach(fn)
}

func (s *Store) BootstrapCluster() error {
	err := s.node.BootstrapCluster()
	if errors.Is(err, raft.ErrCantBootstrap) {
		s.logger.Info(context.Background(), "Cluster is already bootstrapped, skipping")
		return nil
	}

	return err
}

func (s *Store) IsLeader() bool {
	return s.node.IsLeader()
}

func (s *Store) applyCommand(cmd writeCommand) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	future := s.node.Apply(data)
	if err := future.Error(); err != nil {
		if errors.Is(err, raft.ErrNotLeader) || errors.Is(err, raft.ErrLeadershipLost) {
			return s.node.notLeaderError()
		}
		return err
	}

	if respErr, ok := future.Response().(error); ok && respErr != nil {
		return respErr
	}

	return nil
}

func (s *Store) Close() error {
	s.logger.Info(context.Background(), "Shutting down Raft node gracefully...")
	nodeErr := s.node.Close()

	s.logger.Info(context.Background(), "Shutting down key-value db gracefully...")
	dbErr := s.db.Close()

	err := errors.Join(nodeErr, dbErr)
	if err == nil {
		s.logger.Info(context.Background(), "Store closed gracefully")
	}
	return err
}
