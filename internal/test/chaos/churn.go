package chaos

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type step struct {
	target string
	value  string
}

func (h *Harness) LeaderChurnChaos(ctx context.Context) error {
	nodes := h.cluster.Nodes()
	key := fmt.Sprintf("chaos-leader-churn-%d", os.Getpid())

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, _ = h.cluster.Node(cleanupCtx, nodes[0], "del", key)
	}()

	log.Println("[chaos] checking that the cluster accepts writes before leader churn")
	if err := h.retry(ctx, func() error {
		_, err := h.cluster.Node(ctx, nodes[0], "set", key, "baseline")

		return err
	}); err != nil {
		return err
	}

	steps := make([]step, 0, len(nodes)+1)
	for _, node := range nodes {
		steps = append(steps, step{target: node, value: "committed-after-transfer-to-" + node})
	}
	steps = append(steps, step{target: nodes[0], value: "committed-after-transfer-back-to-" + nodes[0]})

	for _, step := range steps {
		log.Printf("[chaos] transferring leadership to %s\n", step.target)
		if _, err := h.cluster.Node(ctx, nodes[0], "move-leader", step.target); err != nil {
			return err
		}

		log.Printf("[chaos] writing after transfer to %s\n", step.target)
		if err := h.retry(ctx, func() error {
			_, err := h.cluster.Node(ctx, nodes[1], "set", key, step.value)

			return err
		}); err != nil {
			return err
		}

		for _, node := range nodes {
			actual, err := h.cluster.Node(ctx, node, "get", key)
			if err != nil {
				return err
			}
			if strings.TrimSpace(actual) != step.value {
				return fmt.Errorf("%s returned %q, want %q", node, actual, step.value)
			}
		}
	}

	log.Println("[chaos] PASS: leadership moved across all nodes and writes remained available")

	return nil
}
