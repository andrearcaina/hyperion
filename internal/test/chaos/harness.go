package chaos

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Harness struct {
	cluster  Cluster
	attempts int
	delay    time.Duration
}

type Cluster interface {
	Nodes() []string
	Node(context.Context, string, ...string) (string, error)
	Kill(context.Context, string) error
	Restart(context.Context, string) error
	Partition(context.Context, string) error
	Heal(context.Context, string) error
}

func otherNodes(nodes []string, excluded string) []string {
	result := make([]string, 0, len(nodes)-1)
	for _, node := range nodes {
		if node != excluded {
			result = append(result, node)
		}
	}
	return result
}

func NewHarness(clusterName string, override ...Cluster) (*Harness, error) {
	backend := Cluster(nil)
	if len(override) == 0 {
		switch strings.ToLower(clusterName) {
		case "docker":
			backend = newDockerCluster()
		case "k8s", "kubernetes":
			backend = newKubernetesCluster("hyperion")
		default:
			return nil, fmt.Errorf("unknown cluster %q (want docker or k8s)", clusterName)
		}
	} else {
		backend = override[0]
	}

	return &Harness{
		cluster:  backend,
		attempts: 30,
		delay:    time.Second,
	}, nil
}

func (h *Harness) retry(ctx context.Context, action func() error) error {
	var err error
	for range h.attempts {
		if err = action(); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(h.delay):
		}
	}

	return err
}

func (h *Harness) write(ctx context.Context, nodes []string, key, value string) error {
	var lastErr error
	for _, node := range nodes {
		_, err := h.cluster.Node(ctx, node, "set", key, value)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return lastErr
}
