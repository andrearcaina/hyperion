package chaos

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andrearcaina/hyperion/internal/client"
	"golang.org/x/sync/errgroup"
)

const concurrentWriteClients = 12

var concurrentWriteAddresses = []string{
	"http://127.0.0.1:8080",
	"http://127.0.0.1:8082",
	"http://127.0.0.1:8084",
}

func (h *Harness) ConcurrentWritesChaos(ctx context.Context) error {
	keys := make([]string, concurrentWriteClients)
	values := make([]string, concurrentWriteClients)
	clients := make([]client.Client, concurrentWriteClients)

	for clientID := range concurrentWriteClients {
		keys[clientID] = fmt.Sprintf("chaos-concurrent-write-%d-%d", os.Getpid(), clientID)
		values[clientID] = fmt.Sprintf("client-%d", clientID)
		clients[clientID] = client.NewHTTP(concurrentWriteAddresses[clientID%len(concurrentWriteAddresses)], 5*time.Second)
	}

	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for _, key := range keys {
			_ = clients[0].Delete(cleanupCtx, key)
		}

		for _, client := range clients {
			_ = client.Close()
		}
	}()

	log.Printf("[chaos] starting %d concurrent clients and writing 10 keys each\n", concurrentWriteClients)
	group, groupCtx := errgroup.WithContext(ctx)
	for clientID := range concurrentWriteClients {
		group.Go(func() error {
			if err := h.retry(groupCtx, func() error {
				_, err := clients[clientID].Put(groupCtx, keys[clientID], []byte(values[clientID]))

				return err
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
		reader := clients[(clientID+1)%len(clients)]
		var actual []byte

		if err := h.retry(ctx, func() error {
			entry, err := reader.Get(ctx, key)
			if err == nil {
				actual = entry.Value
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
