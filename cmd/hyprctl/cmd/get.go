package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a key or list all keys",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			key := args[0]

			entry, err := hyprClient.Get(cmd.Context(), key)
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(cmd, entry)
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(entry.Value))
		} else if len(args) == 0 {
			entries, err := hyprClient.List(cmd.Context())
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(cmd, entries)
			}

			for _, entry := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", entry.Key, entry.Value)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
