package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/hashicorp/raft"
)

type Store struct {
	db     *db.DB
	node   *Node
	logger *logger.Logger
}

func New(db *db.DB, logger *logger.Logger, cfg *NodeConfig) (*Store, error) {
	fsm, err := NewFSM(db, logger)
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
	return s.applyCommand(writeCommand{
		Op:    commandSet,
		Key:   key,
		Value: value,
	})
}

func (s *Store) Delete(key string) error {
	return s.applyCommand(writeCommand{
		Op:  commandDelete,
		Key: key,
	})
}

func (s *Store) Join(nodeID, nodeAddress string) error {
	return s.node.Join(nodeID, nodeAddress)
}

func (s *Store) Get(key string) ([]byte, error) {
	return s.db.Get([]byte(key))
}

func (s *Store) ForEach(fn func(key, value []byte) error) error {
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
	return future.Error()
}

func (s *Store) Close() error {
	s.logger.Info(context.Background(), "Shutting down Raft node gracefully...")
	if err := s.node.Close(); err != nil {
		return err
	}

	s.logger.Info(context.Background(), "Shutting down key-value db gracefully...")
	if err := s.db.Close(); err != nil {
		return err
	}

	s.logger.Info(context.Background(), "Store closed gracefully")
	return nil
}
