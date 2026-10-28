//go:build integration

package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/andrearcaina/hyperion/internal/client"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const integrationImage = "hyperion-integration:test"

var (
	integrationImageOnce sync.Once
	integrationImageErr  error
)

const (
	containerHTTPPort = "8080/tcp"
	containerGRPCPort = "8081/tcp"
	clientTimeout     = 10 * time.Second
)

type node struct {
	http *client.HTTPClient
	grpc *client.GRPCClient
}

func ensureIntegrationImage(ctx context.Context) error {
	integrationImageOnce.Do(func() {
		imageBuilder, err := testcontainers.Run(ctx, "",
			testcontainers.WithDockerfile(testcontainers.FromDockerfile{
				Context:    "../../../",
				Dockerfile: "Dockerfile",
				Repo:       "hyperion-integration",
				Tag:        "test",
				KeepImage:  true,
			}),
			testcontainers.WithNoStart(),
		)
		if err != nil {
			integrationImageErr = err
			return
		}
		integrationImageErr = imageBuilder.Terminate(context.Background())
	})
	return integrationImageErr
}

func TestNodeHTTPAndGRPC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := ensureIntegrationImage(ctx); err != nil {
		t.Fatalf("build integration image: %v", err)
	}

	n := startNode(t, ctx)
	defer n.http.Close()
	defer n.grpc.Close()

	t.Run("HTTP write is readable through gRPC", func(t *testing.T) {
		entry, err := n.http.Put(ctx, "integration-http", []byte("hello from HTTP"))
		if err != nil {
			t.Fatalf("HTTP put: %v", err)
		}
		if string(entry.Value) != "hello from HTTP" {
			t.Fatalf("HTTP put returned %q", entry.Value)
		}

		entry, err = n.grpc.Get(ctx, "integration-http")
		if err != nil {
			t.Fatalf("gRPC get: %v", err)
		}
		if string(entry.Value) != "hello from HTTP" {
			t.Fatalf("gRPC get returned %q", entry.Value)
		}
	})

	t.Run("gRPC write is listed and deleted through HTTP", func(t *testing.T) {
		if _, err := n.grpc.Put(ctx, "integration-grpc", []byte("hello from gRPC")); err != nil {
			t.Fatalf("gRPC put: %v", err)
		}

		entries, err := n.http.List(ctx)
		if err != nil {
			t.Fatalf("HTTP list: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("HTTP list returned %d entries, want 2", len(entries))
		}

		if err := n.http.Delete(ctx, "integration-http"); err != nil {
			t.Fatalf("HTTP delete: %v", err)
		}
		if _, err := n.grpc.Get(ctx, "integration-http"); err == nil {
			t.Fatal("gRPC get succeeded after delete")
		}
	})
}

func startNode(t *testing.T, ctx context.Context) *node {
	t.Helper()
	container, err := testcontainers.Run(ctx, integrationImage,
		testcontainers.WithExposedPorts(containerHTTPPort, containerGRPCPort),
		testcontainers.WithCmd("hyprd", "--node-id", "integration-node", "--node-addr", "127.0.0.1:9001", "--bootstrap"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/healthz").
			WithPort(containerHTTPPort).
			WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		t.Fatalf("start Hyperion container: %v", err)
	}
	testcontainers.CleanupContainer(t, container)

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("resolve container host: %v", err)
	}
	httpPort, err := container.MappedPort(ctx, containerHTTPPort)
	if err != nil {
		t.Fatalf("resolve HTTP port: %v", err)
	}
	grpcPort, err := container.MappedPort(ctx, containerGRPCPort)
	if err != nil {
		t.Fatalf("resolve gRPC port: %v", err)
	}

	httpClient := client.NewHTTPClient(
		fmt.Sprintf("http://%s:%s", host, httpPort.Port()),
		clientTimeout,
	)

	grpcClient, err := client.NewGRPCClient(
		fmt.Sprintf("%s:%s", host, grpcPort.Port()),
		clientTimeout,
	)
	if err != nil {
		httpClient.Close()
		t.Fatalf("create gRPC client: %v", err)
	}

	return &node{
		http: httpClient,
		grpc: grpcClient,
	}
}
