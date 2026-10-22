//go:build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/andrearcaina/hyperion/internal/test/chaos"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type fiveNodeCluster struct {
	network    *testcontainers.DockerNetwork
	containers map[string]testcontainers.Container
}

func TestFiveNodeClusterAndChaos(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := ensureIntegrationImage(ctx); err != nil {
		t.Fatalf("build integration image: %v", err)
	}

	cluster := startFiveNodeCluster(t, ctx)
	nodes := cluster.Nodes()

	if _, err := cluster.Node(ctx, nodes[0], "set", "integration-five-node", "hello-from-five-nodes"); err != nil {
		t.Fatalf("write through first node: %v", err)
	}
	for _, node := range nodes {
		value, err := cluster.Node(ctx, node, "get", "integration-five-node")
		if err != nil {
			t.Fatalf("read through %s: %v", node, err)
		}
		if strings.TrimSpace(value) != "hello-from-five-nodes" {
			t.Fatalf("read through %s = %q", node, value)
		}
	}

	harness, err := chaos.NewHarness("testcontainers", cluster)
	if err != nil {
		t.Fatalf("create chaos harness: %v", err)
	}
	for _, scenario := range []struct {
		name string
		fn   func(context.Context) error
	}{
		{"network-partition", harness.NetworkPartitionChaos},
		{"sigkill", harness.KillChaos},
		{"concurrent-writes", harness.ConcurrentWritesChaos},
		{"leader-churn", harness.LeaderChurnChaos},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if err := scenario.fn(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func startFiveNodeCluster(t *testing.T, ctx context.Context) *fiveNodeCluster {
	t.Helper()

	nw, err := network.New(ctx)
	if err != nil {
		t.Fatalf("create test network: %v", err)
	}
	t.Cleanup(func() { _ = nw.Remove(context.Background()) })

	cluster := &fiveNodeCluster{
		network:    nw,
		containers: make(map[string]testcontainers.Container, 5),
	}

	for i := 1; i <= 5; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		args := []string{"hyprd", "--node-id", nodeID, "--node-addr", nodeID + ":9001"}
		if i == 1 {
			args = append(args, "--bootstrap")
		}

		container, err := testcontainers.Run(ctx, integrationImage,
			testcontainers.WithExposedPorts("8080/tcp"),
			testcontainers.WithCmd(args...),
			testcontainers.WithWaitStrategy(wait.ForHTTP("/healthz").
				WithPort("8080/tcp").
				WithStartupTimeout(90*time.Second)),
			network.WithNetwork([]string{nodeID}, nw),
		)
		if err != nil {
			t.Fatalf("start %s: %v", nodeID, err)
		}
		testcontainers.CleanupContainer(t, container)
		cluster.containers[nodeID] = container
	}

	for i := 2; i <= 5; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		if err := retry(ctx, 30, time.Second, func() error {
			_, err := cluster.Node(ctx, "node-1", "join", "--node-id", nodeID, "--node-addr", nodeID+":9001")
			return err
		}); err != nil {
			t.Fatalf("join %s: %v", nodeID, err)
		}
	}

	return cluster
}

func (c *fiveNodeCluster) Nodes() []string {
	nodes := make([]string, 5)
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("node-%d", i)
		nodes[i-1] = id
	}
	return nodes
}

func (c *fiveNodeCluster) Node(ctx context.Context, nodeID string, args ...string) (string, error) {
	container, ok := c.containers[nodeID]
	if !ok || !container.IsRunning() {
		return "", fmt.Errorf("node %s is not running", nodeID)
	}
	command := append([]string{"hyprctl", "--addr", "http://127.0.0.1:8080", "--timeout", "2s"}, args...)
	return c.exec(ctx, container, command...)
}

func (c *fiveNodeCluster) Kill(ctx context.Context, nodeID string) error {
	container := c.containers[nodeID]
	return runDocker(ctx, "kill", container.GetContainerID())
}

func (c *fiveNodeCluster) Restart(ctx context.Context, nodeID string) error {
	return c.containers[nodeID].Start(ctx)
}

func (c *fiveNodeCluster) Partition(ctx context.Context, nodeID string) error {
	if err := runDocker(ctx, "network", "disconnect", c.network.Name, c.containers[nodeID].GetContainerID()); err != nil {
		return err
	}
	return nil
}

func (c *fiveNodeCluster) Heal(ctx context.Context, nodeID string) error {
	if err := runDocker(ctx, "network", "connect", "--alias", nodeID, c.network.Name, c.containers[nodeID].GetContainerID()); err != nil {
		return err
	}
	return nil
}

func (c *fiveNodeCluster) exec(ctx context.Context, container testcontainers.Container, args ...string) (string, error) {
	exitCode, output, err := container.Exec(ctx, args, tcexec.Multiplexed())
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(output)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", fmt.Errorf("command %q exited with %d: %s", args[0], exitCode, strings.TrimSpace(string(data)))
	}
	return strings.TrimSpace(string(data)), nil
}

func runDocker(ctx context.Context, args ...string) error {
	command := exec.CommandContext(ctx, "docker", args...)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return fmt.Errorf("docker %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(output.String()))
	}
	return nil
}

func retry(ctx context.Context, attempts int, delay time.Duration, action func() error) error {
	var err error
	for range attempts {
		if err = action(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}
