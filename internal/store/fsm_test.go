package store

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/hashicorp/raft"
)

func TestFSMApplyAndRestoreSnapshot(t *testing.T) {
	sourceDB := openTestDB(t, "source")
	fsm, err := NewFSM(sourceDB)
	if err != nil {
		t.Fatal(err)
	}

	apply := func(command string) {
		t.Helper()
		response := fsm.Apply(&raft.Log{Data: []byte(command)})
		if err, ok := response.(error); ok && err != nil {
			t.Fatal(err)
		}
	}

	apply(`{"op":"set","key":"animal","value":"eWFr"}`)
	apply(`{"op":"set","key":"colour","value":"Ymx1ZQ=="}`)
	apply(`{"op":"delete","key":"missing"}`)

	snapshot, err := fsm.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	data := snapshot.(*FSMSnapshot).data

	destinationDB := openTestDB(t, "destination")
	destinationFSM, err := NewFSM(destinationDB)
	if err != nil {
		t.Fatal(err)
	}
	if err := destinationFSM.Restore(io.NopCloser(bytes.NewReader(data))); err != nil {
		t.Fatal(err)
	}

	value, err := destinationDB.Get([]byte("animal"))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := string(value), "yak"; got != want {
		t.Fatalf("restored value = %q, want %q", got, want)
	}
}

func TestFSMRejectsUnknownCommand(t *testing.T) {
	fsm, err := NewFSM(openTestDB(t, "db"))
	if err != nil {
		t.Fatal(err)
	}

	response := fsm.Apply(&raft.Log{Data: []byte(`{"op":"surprise","key":"k"}`)})
	if response == nil {
		t.Fatal("unknown command unexpectedly succeeded")
	}
}

func openTestDB(t *testing.T, name string) *db.DB {
	t.Helper()

	database, err := db.New(filepath.Join(t.TempDir(), name))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = database.Close() })
	return database
}
