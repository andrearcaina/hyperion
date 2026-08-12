package chaos

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func (h *Harness) NetworkPartitionChaos(ctx context.Context) error {
	key := fmt.Sprintf("chaos-network-partition-%d", os.Getpid())
	value := fmt.Sprintf("committed-by-majority-%d", os.Getpid())

	container, err := h.docker.container(ctx, "node-1")
	if err != nil {
		return err
	}
	if container == "" {
		return errors.New("node-1 is not running; run 'make docker-run' first")
	}

	network, err := h.docker.network(ctx, container)
	if err != nil {
		return err
	}

	partitioned := false
	defer func() {
		if partitioned {
			fmt.Printf("[chaos] reconnecting node-1 to %s\n", network)
			if err := h.docker.connect(context.Background(), network, container, "node-1"); err != nil {
				fmt.Fprintln(os.Stderr, "[chaos] cleanup failed:", err)
			}
		}
		_, _ = h.docker.node(context.Background(), "node-2", "del", key)
	}()

	fmt.Println("[chaos] checking that the cluster accepts writes before injecting a fault")
	if err := h.retry(ctx, func() error {
		_, err := h.docker.node(ctx, "node-2", "set", key, "baseline")
		return err
	}); err != nil {
		return err
	}

	fmt.Printf("[chaos] disconnecting node-1 from %s\n", network)
	if err := h.docker.disconnect(ctx, network, container); err != nil {
		return err
	}
	partitioned = true
	time.Sleep(5 * time.Second)

	fmt.Println("[chaos] checking that isolated node-1 cannot commit")
	if _, err := h.docker.node(ctx, "node-1", "set", key, "should-not-commit"); err == nil {
		return errors.New("node-1 unexpectedly committed a write while isolated")
	}

	fmt.Println("[chaos] writing through the majority via node-2")
	if err := h.retry(ctx, func() error {
		_, err := h.docker.node(ctx, "node-2", "set", key, value)
		return err
	}); err != nil {
		return err
	}

	fmt.Println("[chaos] reading the committed value through node-3")
	actual, err := h.docker.node(ctx, "node-3", "get", key)
	if err != nil {
		return err
	}
	if actual != value {
		return fmt.Errorf("node-3 returned %q, want %q", actual, value)
	}

	fmt.Printf("[chaos] reconnecting node-1 to %s\n", network)
	if err := h.docker.connect(ctx, network, container, "node-1"); err != nil {
		return err
	}
	partitioned = false

	fmt.Println("[chaos] waiting for node-1 to catch up")
	if err := h.retry(ctx, func() error {
		actual, err := h.docker.node(ctx, "node-1", "get", key)
		if err == nil && strings.TrimSpace(actual) != value {
			err = fmt.Errorf("node-1 returned %q, want %q", actual, value)
		}
		return err
	}); err != nil {
		return err
	}

	fmt.Println("[chaos] PASS: the majority stayed available and the isolated node recovered")
	return nil
}
