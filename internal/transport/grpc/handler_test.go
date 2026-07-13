package grpc

import (
	"context"
	"errors"
	"net"
	"sort"
	"testing"

	"github.com/andrearcaina/hyperion/internal/store"
	hyperionv1 "github.com/andrearcaina/hyperion/proto"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type memoryStore struct {
	values map[string][]byte
}

func (s *memoryStore) Set(key string, value []byte) error {
	if key == "" {
		return store.ErrInvalidKey
	}

	s.values[key] = append([]byte(nil), value...)
	return nil
}

func (s *memoryStore) Get(key string) ([]byte, error) {
	value, ok := s.values[key]
	if !ok {
		return nil, store.ErrNotFound
	}

	return append([]byte(nil), value...), nil
}

func (s *memoryStore) Delete(key string) error { delete(s.values, key); return nil }
func (s *memoryStore) Join(_, _ string) error  { return nil }
func (s *memoryStore) ForEach(fn func([]byte, []byte) error) error {
	keys := make([]string, 0, len(s.values))
	for key := range s.values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		if err := fn([]byte(key), s.values[key]); err != nil {
			return err
		}
	}

	return nil
}

func TestHandlerRoundTrip(t *testing.T) {
	client := newTestClient(t, &memoryStore{values: make(map[string][]byte)})
	ctx := context.Background()

	if _, err := client.Put(ctx, &hyperionv1.PutRequest{Key: "hello", Value: []byte("world")}); err != nil {
		t.Fatal(err)
	}

	got, err := client.Get(ctx, &hyperionv1.GetRequest{Key: "hello"})
	if err != nil {
		t.Fatal(err)
	}

	if string(got.Entry.Value) != "world" {
		t.Fatalf("value = %q, want world", got.Entry.Value)
	}

	if _, err := client.Delete(ctx, &hyperionv1.DeleteRequest{Key: "hello"}); err != nil {
		t.Fatal(err)
	}

	_, err = client.Get(ctx, &hyperionv1.GetRequest{Key: "hello"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("Get error code = %s, want NotFound", status.Code(err))
	}
}

func TestGRPCErrorMapping(t *testing.T) {
	if code := status.Code(grpcError(store.ErrInvalidKey)); code != codes.InvalidArgument {
		t.Fatalf("invalid key code = %s", code)
	}

	if code := status.Code(grpcError(&store.NotLeaderError{NodeID: "n2"})); code != codes.FailedPrecondition {
		t.Fatalf("not leader code = %s", code)
	}

	if code := status.Code(grpcError(errors.New("disk failed"))); code != codes.Internal {
		t.Fatalf("internal error code = %s", code)
	}
}

func newTestClient(t *testing.T, st Store) hyperionv1.HyperionClient {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := googlegrpc.NewServer()
	hyperionv1.RegisterHyperionServer(server, NewHandler(st))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	connection, err := googlegrpc.NewClient(
		"passthrough:///bufnet",
		googlegrpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = connection.Close() })

	return hyperionv1.NewHyperionClient(connection)
}
