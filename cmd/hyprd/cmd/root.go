package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/andrearcaina/hyperion/internal/server"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	port string
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

		dbPath := filepath.Join(home, ".hyperion", "data", port)
		if err := os.MkdirAll(dbPath, 0755); err != nil {
			return err
		}

		cfg := &server.ServerConfig{
			Port:   port,
			DBPath: dbPath,
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
			log.Printf("Server exited with error: %v", err)
			return err
		}

		log.Println("Server exited cleanly")
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
}
