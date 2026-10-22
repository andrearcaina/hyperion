package chaos

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"time"
)

func (h *Harness) NetworkPartitionChaos(ctx context.Context) error {
	nodes := h.cluster.Nodes()
	partitionedService := nodes[rand.IntN(len(nodes))]
	remainingServices := otherNodes(nodes, partitionedService)

	key := fmt.Sprintf("chaos-network-partition-%d", os.Getpid())
	value := fmt.Sprintf("committed-by-majority-%d", os.Getpid())

	partitioned := false
	defer func() {
		if partitioned {
			log.Printf("[chaos] healing partition for %s\n", partitionedService)

			if err := h.cluster.Heal(context.Background(), partitionedService); err != nil {
				fmt.Fprintln(os.Stderr, "[chaos] cleanup failed:", err)
			}
		}

		_, _ = h.cluster.Node(context.Background(), remainingServices[0], "del", key)
	}()

	log.Println("[chaos] checking that the cluster accepts writes before injecting a fault")
	if err := h.retry(ctx, func() error {
		return h.write(ctx, remainingServices, key, "baseline")
	}); err != nil {
		return err
	}

	log.Printf("[chaos] partitioning %s\n", partitionedService)
	if err := h.cluster.Partition(ctx, partitionedService); err != nil {
		return err
	}
	partitioned = true
	time.Sleep(5 * time.Second)

	log.Printf("[chaos] checking that isolated %s cannot commit\n", partitionedService)
	if _, err := h.cluster.Node(ctx, partitionedService, "set", key, "should-not-commit"); err == nil {
		return fmt.Errorf("%s unexpectedly committed a write while isolated", partitionedService)
	}

	log.Printf("[chaos] writing through the majority via %s\n", remainingServices[0])
	if err := h.retry(ctx, func() error {
		return h.write(ctx, remainingServices, key, value)
	}); err != nil {
		return err
	}

	log.Printf("[chaos] reading the committed value through %s\n", remainingServices[1])
	actual, err := h.cluster.Node(ctx, remainingServices[1], "get", key)
	if err != nil {
		return err
	}
	if actual != value {
		return fmt.Errorf("%s returned %q, want %q", remainingServices[1], actual, value)
	}

	log.Printf("[chaos] healing partition for %s\n", partitionedService)
	if err := h.cluster.Heal(ctx, partitionedService); err != nil {
		return err
	}
	partitioned = false

	log.Printf("[chaos] waiting for %s to catch up\n", partitionedService)
	if err := h.retry(ctx, func() error {
		actual, err := h.cluster.Node(ctx, partitionedService, "get", key)
		if err == nil && strings.TrimSpace(actual) != value {
			err = fmt.Errorf("%s returned %q, want %q", partitionedService, actual, value)
		}

		return err
	}); err != nil {
		return err
	}

	log.Printf("[chaos] PASS: the majority stayed available and %s recovered\n", partitionedService)

	return nil
}
