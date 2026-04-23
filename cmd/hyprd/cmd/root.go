package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/andrearcaina/hyperion/internal/server"
	"github.com/andrearcaina/hyperion/internal/store"
	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	port        string
	nodeID      string
	nodeTimeout int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hyprd",
	Short: "Start a Hyperion key-value store server",
	Long: `Start a Hyperion key-value store server that uses BadgerDB under the hood.

Will be a distributed system later on with Raft Consensus Algorithm.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		stPath := filepath.Join(home, ".hyperion", "data", nodeID)
		dbPath := filepath.Join(stPath, "kv")
		nodePath := filepath.Join(stPath, "raft")

		for _, path := range []string{stPath, dbPath, nodePath} {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		}

		logger := logger.New(nil)

		db, err := db.New(dbPath)
		if err != nil {
			return err
		}

		store, err := store.New(db, logger, &store.NodeConfig{
			NodeID:       nodeID,
			NodePath:     nodePath, // path for Raft log and state machine snapshots
			ApplyTimeout: time.Duration(nodeTimeout) * time.Second,
		})
		if err != nil {
			return err
		}

		handler := http2.NewHandler(store, logger)

		srv, err := server.NewServer(port, logger, handler)
		if err != nil {
			return err
		}

		// use errgroup to manage the lifecycle of the server and handle graceful shutdown
		g, ctx := errgroup.WithContext(ctx)

		// start server
		g.Go(func() error {
			return srv.Run()
		})

		// wait for shutdown signal
		g.Go(func() error {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := srv.Close(shutdownCtx); err != nil {
				return err
			}

			if err := store.Close(); err != nil {
				return err
			}

			return nil
		})

		// wait for everything
		if err := g.Wait(); err != nil {
			logger.Error(context.Background(), "Server exited with error", "error", err)
			return err
		}

		logger.Info(context.Background(), "hyprd exited cleanly")
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", ":8080", "Port to listen on")
	rootCmd.Flags().StringVarP(&nodeID, "node", "n", "node-1", "Node ID")
	rootCmd.Flags().IntVarP(&nodeTimeout, "timeout", "t", 5, "Node timeout in seconds")
}
