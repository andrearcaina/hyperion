package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/andrearcaina/hyperion/internal/test/chaos"
)

type scenario struct {
	name string
	fn   func(context.Context) error
}

func main() {
	clusterName := flag.String("cluster", "docker", "cluster backend: docker or k8s")
	flag.Parse()

	input := "all"

	if flag.NArg() == 1 {
		input = flag.Arg(0)
	} else if flag.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "[chaos] ERROR: too many arguments")
		fmt.Fprintln(os.Stderr, "Usage: hyprchaos [--cluster docker|k8s] [scenario]")
		os.Exit(1)
	}

	if err := runChaos(*clusterName, input); err != nil {
		fmt.Fprintln(os.Stderr, "[chaos] FAIL:", err)
		os.Exit(1)
	}
}

func runChaos(clusterName, input string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	h, err := chaos.NewHarness(clusterName)
	if err != nil {
		return err
	}
	scenarios := []scenario{
		{"network-partition", h.NetworkPartitionChaos},
		{"sigkill", h.KillChaos},
		{"concurrent-writes", h.ConcurrentWritesChaos},
		{"leader-churn", h.LeaderChurnChaos},
	}

	switch input {
	case "all":
		for _, scenario := range scenarios {
			fmt.Printf("---- running scenario: %s ----\n", scenario.name)

			if err := scenario.fn(ctx); err != nil {
				return fmt.Errorf("scenario %q failed: %w", scenario.name, err)
			}
		}

		return nil
	default:
		if input == "network" || input == "partition" {
			input = "network-partition"
		}

		if input == "kill" || input == "kill-9" {
			input = "sigkill"
		}

		if input == "leader" || input == "churn" {
			input = "leader-churn"
		}

		if input == "concurrent" || input == "writes" {
			input = "concurrent-writes"
		}

		for _, scenario := range scenarios {
			if scenario.name == input {
				fmt.Printf("---- running scenario: %s ----\n", scenario.name)

				if err := scenario.fn(ctx); err != nil {
					return fmt.Errorf("scenario %q failed: %w", scenario.name, err)
				}

				return nil
			}
		}

		return fmt.Errorf("unknown input %q", input)
	}
}
