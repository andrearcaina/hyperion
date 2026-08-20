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
	nodes := []string{"node-1", "node-2", "node-3"}
	killedService := nodes[rand.IntN(len(nodes))]
	remainingServices := make([]string, 0, len(nodes)-1)

	for _, node := range nodes {
		if node != killedService {
			remainingServices = append(remainingServices, node)
		}
	}

	key := fmt.Sprintf("chaos-sigkill-%d", os.Getpid())
	value := fmt.Sprintf("committed-after-sigkill-%d", os.Getpid())

	container, err := h.docker.container(ctx, killedService)
	if err != nil {
		return err
	}
	if container == "" {
		return fmt.Errorf("%s is not running; run 'make docker-run' first", killedService)
	}

	restarted := false
	defer func() {
		if !restarted {
			log.Printf("[chaos] restarting %s after SIGKILL\n", killedService)

			if err := h.docker.start(context.Background(), killedService); err != nil {
				fmt.Fprintln(os.Stderr, "[chaos] cleanup failed:", err)
			}
		}

		_, _ = h.docker.node(context.Background(), remainingServices[0], "del", key)
	}()

	log.Println("[chaos] checking that the cluster accepts writes before injecting a fault")
	if err := h.retry(ctx, func() error {
		_, err := h.docker.node(ctx, remainingServices[0], "set", key, "baseline")

		return err
	}); err != nil {
		return err
	}

	log.Printf("[chaos] sending SIGKILL to %s\n", killedService)
	if err := h.docker.kill(ctx, killedService); err != nil {
		return err
	}

	log.Printf("[chaos] writing through the remaining majority via %s\n", remainingServices[0])
	if err := h.retry(ctx, func() error {
		_, err := h.docker.node(ctx, remainingServices[0], "set", key, value)

		return err
	}); err != nil {
		return err
	}

	log.Printf("[chaos] reading the committed value through %s\n", remainingServices[1])
	actual, err := h.docker.node(ctx, remainingServices[1], "get", key)
	if err != nil {
		return err
	}
	if strings.TrimSpace(actual) != value {
		return fmt.Errorf("%s returned %q, want %q", remainingServices[1], actual, value)
	}

	log.Printf("[chaos] restarting %s\n", killedService)
	if err := h.docker.start(ctx, killedService); err != nil {
		return err
	}
	restarted = true

	log.Printf("[chaos] waiting for %s to catch up\n", killedService)
	if err := h.retry(ctx, func() error {
		actual, err := h.docker.node(ctx, killedService, "get", key)
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
