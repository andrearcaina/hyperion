package store

import (
	"encoding/json"
	"fmt"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
)

type Store struct {
	db   *db.DB
	node *Node
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
		db:   db,
		node: node,
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

func (s *Store) Get(key string) ([]byte, error) {
	if !s.node.IsLeader() {
		return nil, fmt.Errorf("node %s is not the leader", s.node.GetID())
	}

	return s.db.Get([]byte(key))
}

func (s *Store) ForEach(fn func(key, value []byte) error) error {
	if !s.node.IsLeader() {
		return fmt.Errorf("node %s is not the leader", s.node.GetID())
	}

	return s.db.ForEach(fn)
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
	if err := s.db.Close(); err != nil {
		return err
	}

	if err := s.node.Close(); err != nil {
		return err
	}

	return nil
}
