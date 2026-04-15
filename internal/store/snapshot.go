package store

import (
	"encoding/json"

	"github.com/hashicorp/raft"
)

type FSMSnapshot struct {
	data map[string][]byte
}

func (s *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	if err := json.NewEncoder(sink).Encode(s.data); err != nil {
		_ = sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *FSMSnapshot) Release() {}
