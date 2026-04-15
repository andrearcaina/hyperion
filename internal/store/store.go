package store

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

type NodeConfig struct {
	NodeID       string
	ApplyTimeout time.Duration
	DBPath       string
}

type Store struct {
	db   *db.DB
	raft *raft.Raft
	cfg  *NodeConfig
}

func New(db *db.DB, logger *logger.Logger, cfg *NodeConfig) (*Store, error) {
	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.NodeID)

	fsm := NewFSM(db, logger)

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

	// bootstraps the current node as the initial leader of the cluster
	if err := r.BootstrapCluster(bootstrap).Error(); err != nil && !errors.Is(err, raft.ErrCantBootstrap) {
		return nil, err
	}

	return &Store{
		db:   db,
		raft: r,
		cfg:  cfg,
	}, nil
}

func (s *Store) Set(key string, value []byte) error {
	return s.applyCommand(command{
		Op:    commandSet,
		Key:   key,
		Value: value,
	})
}

func (s *Store) Delete(key string) error {
	return s.applyCommand(command{
		Op:  commandDelete,
		Key: key,
	})
}

func (s *Store) Get(key string) ([]byte, error) {
	return s.db.Get([]byte(key))
}

func (s *Store) ForEach(fn func(key, value []byte) error) error {
	return s.db.ForEach(fn)
}

func (s *Store) applyCommand(cmd command) error {
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	future := s.raft.Apply(data, s.cfg.ApplyTimeout)
	return future.Error()
}

func (s *Store) Close() error {
	return s.raft.Shutdown().Error()
}
