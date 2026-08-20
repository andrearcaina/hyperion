package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/andrearcaina/hyperion/internal/chaos"
)

func main() {
	scenario := "all"

	if len(os.Args) == 2 {
		scenario = os.Args[1]
	} else if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "[chaos] ERROR: too many arguments")
		fmt.Fprintln(os.Stderr, "Usage: chaos <scenario> (default: all)")
		os.Exit(1)
	}

	if err := runChaos(scenario); err != nil {
		fmt.Fprintln(os.Stderr, "[chaos] FAIL:", err)
		os.Exit(1)
	}
}

func runChaos(scenario string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	h := chaos.NewHarness()
	scenarios := map[string]func(context.Context) error{
		"network-partition": h.NetworkPartitionChaos,
		"sigkill":           h.KillChaos,
	}

	switch scenario {
	case "all":
		for scenario, fn := range scenarios {
			fmt.Printf("---- running scenario: %s ----\n", scenario)

			if err := fn(ctx); err != nil {
				return fmt.Errorf("scenario %q failed: %w", scenario, err)
			}
		}

		return nil
	default:
		if scenario == "partition" {
			scenario = "network-partition"
		}

		if scenario == "kill" || scenario == "kill-9" {
			scenario = "sigkill"
		}

		fn, ok := scenarios[scenario]
		if !ok {
			return fmt.Errorf("unknown scenario %q", scenario)
		}

		fmt.Printf("---- running scenario: %s ----\n", scenario)
		return fn(ctx)
	}
}
