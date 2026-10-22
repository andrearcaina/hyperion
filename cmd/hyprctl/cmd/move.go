package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var moveLeaderCmd = &cobra.Command{
	Use:   "move-leader <node-id>",
	Short: "Transfer Raft leadership to a node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := hyprClient.MoveLeader(cmd.Context(), args[0]); err != nil {
			return fmt.Errorf("move leader: %w", err)
		}

		if jsonOutput {
			return printJSON(cmd, map[string]string{"status": "transferred", "node_id": args[0]})
		}

		fmt.Fprintln(cmd.OutOrStdout(), "transferred")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveLeaderCmd)
}
