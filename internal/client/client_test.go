package client

import (
	"encoding/json"
	"testing"
)

func TestEntryJSONUsesReadableValue(t *testing.T) {
	encoded, err := json.Marshal(Entry{Key: "greeting", Value: []byte("hello")})
	if err != nil {
		t.Fatal(err)
	}

	const want = `{"key":"greeting","value":"hello"}`
	if got := string(encoded); got != want {
		t.Fatalf("JSON = %s, want %s", got, want)
	}
}
