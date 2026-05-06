package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

func printJSON(cmd *cobra.Command, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = cmd.OutOrStdout().Write(append(data, '\n'))
	return err
}
