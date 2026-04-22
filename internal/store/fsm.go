package store

import (
	"bytes"
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

func NewFSM(db *db.DB, logger *logger.Logger) (*FSM, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	return &FSM{
		db:     db,
		logger: logger,
	}, nil
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
	// create a buffer to hold the snapshot data (instead of writing directly to a map)
	buf := &bytes.Buffer{}

	// create a JSON encoder that writes to the buffer
	enc := json.NewEncoder(buf)

	err := s.db.ForEach(func(key, value []byte) error {
		return enc.Encode(struct {
			Key   []byte
			Value []byte
		}{
			Key:   key,
			Value: value,
		})
	})
	if err != nil {
		return nil, err
	}

	// instead of DB -> map -> JSON, we do DB -> JSON directly by using a buffer and encoder
	return &FSMSnapshot{
		data: buf.Bytes(),
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

	if err := s.db.Clear(); err != nil {
		return err
	}

	for key, value := range data {
		if err := s.db.Set([]byte(key), value); err != nil {
			return err
		}
	}

	return nil
}
