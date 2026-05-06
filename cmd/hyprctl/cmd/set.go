package cmd

import (
	"fmt"
	"net/url"

	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
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

		var respBody http2.KVResponse
		var errBody http2.KVResponse

		resp, err := hyprClient.Put(
			"/hypr/kv/"+url.PathEscape(key),
			value,
			&respBody,
			&errBody,
		)
		if err != nil {
			return err
		}

		if resp.IsError() {
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), errBody.Error)
		}

		if jsonOutput {
			return printJSON(cmd, respBody)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", respBody.Key, respBody.Value)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
}
