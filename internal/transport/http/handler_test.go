package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/andrearcaina/hyperion/internal/store"
)

type followerStore struct {
	err error
}

func (s *followerStore) Set(string, []byte) error                  { return s.err }
func (s *followerStore) Get(string) ([]byte, error)                { return nil, s.err }
func (s *followerStore) Delete(string) error                       { return s.err }
func (s *followerStore) ForEach(func([]byte, []byte) error) error  { return s.err }
func (s *followerStore) Join(string, string, string, string) error { return s.err }

func TestFollowerForwardsHTTPPutToLeader(t *testing.T) {
	transport := &recordingTransport{}
	follower := &followerStore{
		err: &store.NotLeaderError{
			NodeID:            "n2",
			LeaderID:          "n1",
			LeaderHTTPAddress: "n1:8080",
		},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/kv/greeting", strings.NewReader("hello"))
	handler := NewHandler(follower, logger.New(nil))
	handler.proxyTransport = transport
	handler.ServeRoutes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if transport.method != http.MethodPut || transport.path != "/kv/greeting" {
		t.Fatalf("forwarded request = %s %s", transport.method, transport.path)
	}
	if transport.body != "hello" {
		t.Fatalf("body = %q, want hello", transport.body)
	}
	if !transport.forwarded {
		t.Fatal("forward marker is missing")
	}
}

type recordingTransport struct {
	method    string
	path      string
	body      string
	forwarded bool
}

func (t *recordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	t.method = r.Method
	t.path = r.URL.Path
	t.body = string(body)
	t.forwarded = r.Header.Get(forwardedRequestHeader) != ""

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}, nil
}

func TestFollowerDoesNotForwardTwice(t *testing.T) {
	follower := &followerStore{
		err: &store.NotLeaderError{
			NodeID:            "n2",
			LeaderID:          "n1",
			LeaderHTTPAddress: "127.0.0.1:1",
		},
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/kv/greeting", nil)
	request.Header.Set(forwardedRequestHeader, "1")
	NewHandler(follower, logger.New(nil)).ServeRoutes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", recorder.Code)
	}
}

type valueStore struct {
	value []byte
}

func (s *valueStore) Set(string, []byte) error                  { return nil }
func (s *valueStore) Get(string) ([]byte, error)                { return s.value, nil }
func (s *valueStore) Delete(string) error                       { return nil }
func (s *valueStore) ForEach(func([]byte, []byte) error) error  { return nil }
func (s *valueStore) Join(string, string, string, string) error { return nil }

func TestGetEncodesBinaryValueAsBase64(t *testing.T) {
	want := []byte{0xff, 0x00}
	handler := NewHandler(
		&valueStore{
			value: want,
		},
		logger.New(nil),
	)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/kv/binary", nil)
	handler.ServeRoutes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var response KVResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	got, err := base64.StdEncoding.DecodeString(response.Value)
	if err != nil {
		t.Fatalf("response value is not valid base64: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("decoded value = %v, want %v", got, want)
	}
}
