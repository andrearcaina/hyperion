package chaos

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/andrearcaina/hyperion/internal/client"
)

type step struct {
	target string
	value  string
}

func (h *Harness) LeaderChurnChaos(ctx context.Context) error {
	nodes := []string{"node-1", "node-2", "node-3"}
	key := fmt.Sprintf("chaos-leader-churn-%d", os.Getpid())
	clients := make([]client.Client, len(concurrentWriteAddresses))

	for i, address := range concurrentWriteAddresses {
		clients[i] = client.NewHTTP(address, 5*time.Second)
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_ = clients[0].Delete(cleanupCtx, key)

		for _, client := range clients {
			_ = client.Close()
		}
	}()

	log.Println("[chaos] checking that the cluster accepts writes before leader churn")
	if err := h.retry(ctx, func() error {
		_, err := h.docker.node(ctx, nodes[0], "set", key, "baseline")

		return err
	}); err != nil {
		return err
	}

	steps := []step{
		{target: "node-1", value: "committed-after-transfer-to-node-1"},
		{target: "node-2", value: "committed-after-transfer-to-node-2"},
		{target: "node-3", value: "committed-after-transfer-to-node-3"},
		{target: "node-1", value: "committed-after-transfer-back-to-node-1"},
	}

	for _, step := range steps {
		log.Printf("[chaos] transferring leadership to %s\n", step.target)
		if err := clients[0].TransferLeadership(ctx, step.target); err != nil {
			return err
		}

		log.Printf("[chaos] writing after transfer to %s\n", step.target)
		if err := h.retry(ctx, func() error {
			_, err := h.docker.node(ctx, nodes[1], "set", key, step.value)

			return err
		}); err != nil {
			return err
		}

		for _, node := range nodes {
			actual, err := h.docker.node(ctx, node, "get", key)
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
