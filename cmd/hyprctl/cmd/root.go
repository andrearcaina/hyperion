/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"time"

	"github.com/andrearcaina/hyperion/internal/client"
	"github.com/spf13/cobra"
)

var (
	serverAddr     string
	requestTimeout time.Duration
	hyprClient     *client.Client
	jsonOutput     bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hyprctl",
	Short: "CLI for interacting with hyprd over HTTP",
	Long: `hyprctl is a CLI wrapper around the hyprd HTTP API.

It can read and write keys, list values, and manage Raft node joins.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		hyprClient = client.New(serverAddr, requestTimeout)
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
	rootCmd.PersistentFlags().StringVarP(&serverAddr, "addr", "a", "http://127.0.0.1:8080", "hyprd HTTP address")
	rootCmd.PersistentFlags().DurationVar(&requestTimeout, "timeout", 5*time.Second, "request timeout")
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "output in JSON format")
}
