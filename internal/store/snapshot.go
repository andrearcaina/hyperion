package store

import (
	"github.com/hashicorp/raft"
)

type FSMSnapshot struct {
	data []byte
}

func (s *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(s.data); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *FSMSnapshot) Release() {}
