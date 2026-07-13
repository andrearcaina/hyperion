package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/andrearcaina/hyperion/internal/client"
	"github.com/spf13/cobra"
)

var (
	serverAddr     string
	protocol       string
	requestTimeout time.Duration
	hyprClient     client.Client
	jsonOutput     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hyprctl",
	Short: "CLI for interacting with hyprd",
	Long: `hyprctl is a CLI for the hyprd HTTP and gRPC APIs.

It can read and write keys, list values, and manage Raft node joins.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		switch protocol {
		case "http":
			if serverAddr == "" {
				serverAddr = "http://127.0.0.1:8080"
			}
			hyprClient = client.NewHTTP(serverAddr, requestTimeout)
		case "grpc":
			if serverAddr == "" {
				serverAddr = "127.0.0.1:8081"
			}
			var err error
			hyprClient, err = client.NewGRPC(serverAddr, requestTimeout)
			return err
		default:
			return fmt.Errorf("unsupported protocol %q (use http or grpc)", protocol)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error { return hyprClient.Close() },
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverAddr, "addr", "a", "", "hyprd API address (defaults to the selected protocol's local port)")
	rootCmd.PersistentFlags().StringVar(&protocol, "protocol", "http", "API protocol: http or grpc")
	rootCmd.PersistentFlags().DurationVar(&requestTimeout, "timeout", 5*time.Second, "request timeout")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
}
