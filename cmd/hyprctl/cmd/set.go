package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a key",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		entry, err := hyprClient.Put(cmd.Context(), key, []byte(value))
		if err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(cmd, entry)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", entry.Key, entry.Value)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
