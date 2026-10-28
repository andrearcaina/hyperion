package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/andrearcaina/hyperion/internal/test/chaos"
	"github.com/spf13/cobra"
)

var (
	clusterName  string
	scenarioName string
)

type scenario struct {
	name string
	fn   func(context.Context) error
}

var rootCmd = &cobra.Command{
	Use:   "hyprchaos",
	Short: "Run Hyperion chaos scenarios",
	Long: `hyprchaos is a CLI for a running cluster of hyprd processes.

It can run 4 scenarios, network partitions, sigkills, concurrent writes, leader-churns, and all.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		h, err := chaos.NewHarness(clusterName)
		if err != nil {
			return fmt.Errorf("[chaos] FAIL: %w", err)
		}
		scenarios := []scenario{
			{"network-partition", h.NetworkPartitionChaos},
			{"sigkill", h.KillChaos},
			{"concurrent-writes", h.ConcurrentWritesChaos},
			{"leader-churn", h.LeaderChurnChaos},
		}

		switch scenarioName {
		case "all":
			for _, scenario := range scenarios {
				fmt.Printf("---- running scenario: %s ----\n", scenario.name)

				if err := scenario.fn(ctx); err != nil {
					return fmt.Errorf("[chaos] FAIL: scenario %q failed: %w", scenario.name, err)
				}
			}
		default:
			if scenarioName == "network" || scenarioName == "partition" {
				scenarioName = "network-partition"
			}

			if scenarioName == "kill" || scenarioName == "kill-9" {
				scenarioName = "sigkill"
			}

			if scenarioName == "leader" || scenarioName == "churn" {
				scenarioName = "leader-churn"
			}

			if scenarioName == "concurrent" || scenarioName == "writes" {
				scenarioName = "concurrent-writes"
			}

			for _, scenario := range scenarios {
				if scenario.name == scenarioName {
					fmt.Printf("---- running scenario: %s ----\n", scenario.name)

					if err := scenario.fn(ctx); err != nil {
						return fmt.Errorf("[chaos] FAIL: scenario %q failed: %w", scenario.name, err)
					}

					return nil
				}
			}

			return fmt.Errorf("[chaos] FAIL: unknown scenario %q", scenarioName)
		}

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
	rootCmd.Flags().StringVarP(&clusterName, "cluster", "c", "docker", "cluster backend: docker or k8s")
	rootCmd.Flags().StringVarP(&scenarioName, "scenario", "s", "all", "chaos scenario to run")
}
