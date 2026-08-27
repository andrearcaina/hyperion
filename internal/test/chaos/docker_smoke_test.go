//go:build smoke

package chaos

import (
	"context"
	"strings"
	"testing"
)

func TestDockerCompose(t *testing.T) {
	ctx := context.Background()
	d := &dockerController{}

	output, err := d.compose(ctx, "ps")
	if err != nil {
		t.Fatalf("compose() failed: %v", err)
	}

	if output == "" {
		t.Fatal("compose() returned empty output")
	}
}

func TestDockerNode(t *testing.T) {
	ctx := context.Background()
	d := &dockerController{}

	output, err := d.node(ctx, "node-1")
	if err != nil {
		t.Fatalf("node() failed: %v", err)
	}

	t.Logf("node output: %s", output)
}

func TestDockerContainer(t *testing.T) {
	ctx := context.Background()
	d := &dockerController{}

	container, err := d.container(ctx, "node-1")
	if err != nil {
		t.Fatalf("container() failed: %v", err)
	}

	if strings.TrimSpace(container) == "" {
		t.Fatal("container() returned empty container ID")
	}

	t.Logf("container ID: %s", container)
}

func TestDockerNetwork(t *testing.T) {
	ctx := context.Background()
	d := &dockerController{}

	container, err := d.container(ctx, "node-1")
	if err != nil {
		t.Fatalf("container() failed: %v", err)
	}

	container = strings.TrimSpace(container)

	network, err := d.network(ctx, container)
	if err != nil {
		t.Fatalf("network() failed: %v", err)
	}

	if network == "" {
		t.Fatal("network() returned empty network")
	}

	t.Logf("network: %s", network)
}

func TestDockerDisconnect(t *testing.T) {
	ctx := context.Background()
	d := &dockerController{}

	container, err := d.container(ctx, "node-1")
	if err != nil {
		t.Fatalf("container() failed: %v", err)
	}

	container = strings.TrimSpace(container)

	network, err := d.network(ctx, container)
	if err != nil {
		t.Fatalf("network() failed: %v", err)
	}

	if err := d.disconnect(ctx, network, container); err != nil {
		t.Fatalf("disconnect() failed: %v", err)
	}

	if err := d.connect(ctx, network, container, "node-1"); err != nil {
		t.Fatalf("failed to reconnect container: %v", err)
	}
}

func TestDockerConnect(t *testing.T) {
	ctx := context.Background()
	d := &dockerController{}

	container, err := d.container(ctx, "node-1")
	if err != nil {
		t.Fatalf("container() failed: %v", err)
	}

	container = strings.TrimSpace(container)

	network, err := d.network(ctx, container)
	if err != nil {
		t.Fatalf("network() failed: %v", err)
	}

	if err := d.disconnect(ctx, network, container); err != nil {
		t.Fatalf("disconnect() failed: %v", err)
	}

	if err := d.connect(ctx, network, container, "node-1"); err != nil {
		t.Fatalf("connect() failed: %v", err)
	}
}
