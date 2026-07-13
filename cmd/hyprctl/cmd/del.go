package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

type deleteResponse struct {
	Status string `json:"status,omitempty"`
}

// delCmd represents the del command
var delCmd = &cobra.Command{
	Use:   "del <key>",
	Short: "Delete a key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if err := hyprClient.Delete(cmd.Context(), key); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}

		if jsonOutput {
			return printJSON(cmd, deleteResponse{Status: "deleted"})
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(delCmd)
}
