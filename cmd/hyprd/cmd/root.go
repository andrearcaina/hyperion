package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/andrearcaina/hyperion/internal/server"
	"github.com/andrearcaina/hyperion/internal/store"
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

		dbPath := filepath.Join(home, ".hyperion", "data", nodeID)
		kvPath := filepath.Join(dbPath, "kv")
		raftPath := filepath.Join(dbPath, "raft")

		for _, path := range []string{dbPath, kvPath, raftPath} {
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		}

		logger := logger.New(nil)

		nodeConfig := &store.NodeConfig{
			NodeID:       nodeID,
			ApplyTimeout: time.Duration(nodeTimeout) * time.Second,
			DBPath:       raftPath, // path for Raft log and state machine snapshots
		}

		cfg := &server.ServerConfig{
			Port:       port,
			DBPath:     kvPath, // path for actual key-value data
			Logger:     logger,
			NodeConfig: nodeConfig,
		}

		srv, err := server.NewServer(cfg)
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
			return srv.Close(shutdownCtx)
		})

		// wait for everything
		if err := g.Wait(); err != nil {
			logger.Error(context.Background(), "Server exited with error", "error", err)
			return err
		}

		logger.Info(context.Background(), "Server exited cleanly")
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
