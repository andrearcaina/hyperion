package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/andrearcaina/hyperion/internal/chaos"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: hyprchaos <scenario>")
		os.Exit(2)
	}

	if err := runChaos(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "[chaos] FAIL:", err)
		os.Exit(1)
	}
}

func runChaos(scenario string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	h := chaos.NewHarness()

	switch scenario {
	case "all":
		fmt.Println("[chaos] not yet implemented for all, will run network-partition for now")
		return h.NetworkPartitionChaos(ctx)
	case "network-partition":
		return h.NetworkPartitionChaos(ctx)
	default:
		return fmt.Errorf("unknown scenario %q", scenario)
	}
}
