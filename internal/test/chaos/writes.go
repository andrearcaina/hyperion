package chaos

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const concurrentWriteClients = 12

func (h *Harness) ConcurrentWritesChaos(ctx context.Context) error {
	nodes := h.cluster.Nodes()
	keys := make([]string, concurrentWriteClients)
	values := make([]string, concurrentWriteClients)

	for clientID := range concurrentWriteClients {
		keys[clientID] = fmt.Sprintf("chaos-concurrent-write-%d-%d", os.Getpid(), clientID)
		values[clientID] = fmt.Sprintf("client-%d", clientID)
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for _, key := range keys {
			_, _ = h.cluster.Node(cleanupCtx, nodes[0], "del", key)
		}
	}()

	log.Printf("[chaos] starting %d concurrent clients and writing 10 keys each\n", concurrentWriteClients)
	group, groupCtx := errgroup.WithContext(ctx)
	for clientID := range concurrentWriteClients {
		group.Go(func() error {
			if err := h.retry(groupCtx, func() error {
				return h.write(groupCtx, nodes, keys[clientID], values[clientID])
			}); err != nil {
				return fmt.Errorf("client %d write: %w", clientID, err)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	log.Println("[chaos] verifying every acknowledged write through another node")
	for clientID, key := range keys {
		var actual []byte

		if err := h.retry(ctx, func() error {
			result, err := h.cluster.Node(ctx, nodes[(clientID+1)%len(nodes)], "get", key)
			if err == nil {
				actual = []byte(strings.TrimSpace(result))
			}

			if err == nil && !bytes.Equal(actual, []byte(values[clientID])) {
				return fmt.Errorf("value = %q, want %q", actual, values[clientID])
			}

			return err
		}); err != nil {
			return fmt.Errorf("verify client %d write: %w", clientID, err)
		}
	}

	log.Printf("[chaos] PASS: %d concurrent client writes committed and replicated\n", concurrentWriteClients)

	return nil
}
