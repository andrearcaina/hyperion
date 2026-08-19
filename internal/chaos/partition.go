package chaos

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

func (h *Harness) NetworkPartitionChaos(ctx context.Context) error {
	nodes := []string{"node-1", "node-2", "node-3"}
	partitionedService := nodes[rand.IntN(len(nodes))]
	remainingServices := make([]string, 0, len(nodes)-1)

	for _, node := range nodes {
		if node != partitionedService {
			remainingServices = append(remainingServices, node)
		}
	}

	key := fmt.Sprintf("chaos-network-partition-%d", os.Getpid())
	value := fmt.Sprintf("committed-by-majority-%d", os.Getpid())

	container, err := h.docker.container(ctx, partitionedService)
	if err != nil {
		return err
	}
	if container == "" {
		return fmt.Errorf("%s is not running; run 'make docker-run' first", partitionedService)
	}

	network, err := h.docker.network(ctx, container)
	if err != nil {
		return err
	}

	partitioned := false
	defer func() {
		if partitioned {
			fmt.Printf("[chaos] reconnecting %s to %s\n", partitionedService, network)
			if err := h.docker.connect(context.Background(), network, container, partitionedService); err != nil {
				fmt.Fprintln(os.Stderr, "[chaos] cleanup failed:", err)
			}
		}
		_, _ = h.docker.node(context.Background(), remainingServices[0], "del", key)
	}()

	fmt.Println("[chaos] checking that the cluster accepts writes before injecting a fault")
	if err := h.retry(ctx, func() error {
		_, err := h.docker.node(ctx, remainingServices[0], "set", key, "baseline")
		return err
	}); err != nil {
		return err
	}

	fmt.Printf("[chaos] disconnecting %s from %s\n", partitionedService, network)
	if err := h.docker.disconnect(ctx, network, container); err != nil {
		return err
	}
	partitioned = true
	time.Sleep(5 * time.Second)

	fmt.Printf("[chaos] checking that isolated %s cannot commit\n", partitionedService)
	if _, err := h.docker.node(ctx, partitionedService, "set", key, "should-not-commit"); err == nil {
		return fmt.Errorf("%s unexpectedly committed a write while isolated", partitionedService)
	}

	fmt.Printf("[chaos] writing through the majority via %s\n", remainingServices[0])
	if err := h.retry(ctx, func() error {
		_, err := h.docker.node(ctx, remainingServices[0], "set", key, value)
		return err
	}); err != nil {
		return err
	}

	fmt.Printf("[chaos] reading the committed value through %s\n", remainingServices[1])
	actual, err := h.docker.node(ctx, remainingServices[1], "get", key)
	if err != nil {
		return err
	}
	if actual != value {
		return fmt.Errorf("%s returned %q, want %q", remainingServices[1], actual, value)
	}

	fmt.Printf("[chaos] reconnecting %s to %s\n", partitionedService, network)
	if err := h.docker.connect(ctx, network, container, partitionedService); err != nil {
		return err
	}
	partitioned = false

	fmt.Printf("[chaos] waiting for %s to catch up\n", partitionedService)
	if err := h.retry(ctx, func() error {
		actual, err := h.docker.node(ctx, partitionedService, "get", key)
		if err == nil && strings.TrimSpace(actual) != value {
			err = fmt.Errorf("%s returned %q, want %q", partitionedService, actual, value)
		}
		return err
	}); err != nil {
		return err
	}

	fmt.Printf("[chaos] PASS: the majority stayed available and %s recovered\n", partitionedService)
	return nil
}
