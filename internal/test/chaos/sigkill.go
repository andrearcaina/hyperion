package chaos

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"
)

func (h *Harness) KillChaos(ctx context.Context) error {
	nodes := h.cluster.Nodes()
	killedService := nodes[rand.IntN(len(nodes))]
	remainingServices := otherNodes(nodes, killedService)

	key := fmt.Sprintf("chaos-sigkill-%d", os.Getpid())
	value := fmt.Sprintf("committed-after-sigkill-%d", os.Getpid())

	restarted := false
	defer func() {
		if !restarted {
			log.Printf("[chaos] restarting %s after SIGKILL\n", killedService)

			if err := h.cluster.Restart(context.Background(), killedService); err != nil {
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

	log.Printf("[chaos] sending SIGKILL to %s\n", killedService)
	if err := h.cluster.Kill(ctx, killedService); err != nil {
		return err
	}

	log.Printf("[chaos] writing through the remaining majority via %s\n", remainingServices[0])
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
	if strings.TrimSpace(actual) != value {
		return fmt.Errorf("%s returned %q, want %q", remainingServices[1], actual, value)
	}

	log.Printf("[chaos] restarting %s\n", killedService)
	if err := h.cluster.Restart(ctx, killedService); err != nil {
		return err
	}
	restarted = true

	log.Printf("[chaos] waiting for %s to catch up\n", killedService)
	if err := h.retry(ctx, func() error {
		actual, err := h.cluster.Node(ctx, killedService, "get", key)
		if err == nil && strings.TrimSpace(actual) != value {
			err = fmt.Errorf("%s returned %q, want %q", killedService, actual, value)
		}

		return err
	}); err != nil {
		return err
	}

	log.Printf("[chaos] PASS: %s recovered after SIGKILL and caught up\n", killedService)

	return nil
}
