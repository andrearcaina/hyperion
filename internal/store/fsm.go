package store

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/hashicorp/raft"
)

type command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value []byte `json:"value,omitempty"`
}

const (
	commandSet    = "set"
	commandDelete = "delete"
)

type FSM struct {
	db     *db.DB
	logger *logger.Logger
}

func NewFSM(db *db.DB, logger *logger.Logger) *FSM {
	return &FSM{
		db:     db,
		logger: logger,
	}
}

func (s *FSM) Apply(log *raft.Log) interface{} {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	switch cmd.Op {
	case commandSet:
		return s.db.Set([]byte(cmd.Key), cmd.Value)
	case commandDelete:
		return s.db.Delete([]byte(cmd.Key))
	default:
		return fmt.Errorf("unknown command: %s", cmd.Op)
	}
}

func (s *FSM) Snapshot() (raft.FSMSnapshot, error) {
	data := map[string][]byte{}

	if err := s.db.ForEach(func(key, value []byte) error {
		valCopy := make([]byte, len(value))
		copy(valCopy, value)
		data[string(key)] = valCopy
		return nil
	}); err != nil {
		return nil, err
	}

	return &FSMSnapshot{
		data: data,
	}, nil
}

func (s *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	payload, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	data := map[string][]byte{}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &data); err != nil {
			return err
		}
	}

	keys := make([][]byte, 0, len(data))
	if err := s.db.ForEach(func(key, _ []byte) error {
		keyCopy := make([]byte, len(key))
		copy(keyCopy, key)
		keys = append(keys, keyCopy)
		return nil
	}); err != nil {
		return err
	}

	for _, key := range keys {
		if err := s.db.Delete(key); err != nil {
			return err
		}
	}

	for key, value := range data {
		if err := s.db.Set([]byte(key), value); err != nil {
			return err
		}
	}

	return nil
}
