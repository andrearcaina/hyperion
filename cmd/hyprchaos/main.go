package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/andrearcaina/hyperion/internal/chaos"
)

type scenario struct {
	name string
	fn   func(context.Context) error
}

func main() {
	input := "all"

	if len(os.Args) == 2 {
		input = os.Args[1]
	} else if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "[chaos] ERROR: too many arguments")
		fmt.Fprintln(os.Stderr, "Usage: chaos <scenario> (default: all)")
		os.Exit(1)
	}

	if err := runChaos(input); err != nil {
		fmt.Fprintln(os.Stderr, "[chaos] FAIL:", err)
		os.Exit(1)
	}
}

func runChaos(input string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	h := chaos.NewHarness()
	scenarios := []scenario{
		{"network-partition", h.NetworkPartitionChaos},
		{"sigkill", h.KillChaos},
		{"concurrent-writes", h.ConcurrentWritesChaos},
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
		if input == "partition" {
			input = "network-partition"
		}

		if input == "kill" || input == "kill-9" {
			input = "sigkill"
		}
		if input == "concurrent" || input == "writes" {
			input = "concurrent-writes"
		}

		for _, scenario := range scenarios {
			if scenario.name == input {
				if err := scenario.fn(ctx); err != nil {
					return fmt.Errorf("scenario %q failed: %w", scenario.name, err)
				}

				return nil
			}
		}

		return fmt.Errorf("unknown input %q", input)
	}
}
