package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestEntryJSONUsesBase64Value(t *testing.T) {
	want := []byte{0xff, 0x00}
	encoded, err := json.Marshal(Entry{Key: "greeting", Value: want})
	if err != nil {
		t.Fatal(err)
	}

	var response struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatal(err)
	}
	got, err := base64.StdEncoding.DecodeString(response.Value)
	if err != nil {
		t.Fatalf("JSON value is not valid base64: %v", err)
	}
	if response.Key != "greeting" || !bytes.Equal(got, want) {
		t.Fatalf("decoded response = %#v, want greeting and %v", response, want)
	}
}

func TestEntryFromHTTPDecodesBinaryValue(t *testing.T) {
	want := []byte{0xff, 0x00}
	entry, err := entryFromHTTP(httpEntry{
		Key:   "binary",
		Value: base64.StdEncoding.EncodeToString(want),
	})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(entry.Value, want) {
		t.Fatalf("value = %v, want %v", entry.Value, want)
	}
}
