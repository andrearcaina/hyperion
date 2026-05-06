package cmd

import (
	"fmt"
	"net/url"

	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
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

			var body http2.KVResponse
			var errBody http2.KVResponse

			resp, err := hyprClient.Get("/hypr/kv/"+url.PathEscape(key), &body, &errBody)
			if err != nil {
				return err
			}

			if resp.IsError() {
				return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), errBody.Error)
			}

			if jsonOutput {
				return printJSON(cmd, body)
			}

			fmt.Fprintln(cmd.OutOrStdout(), body.Value)
		} else if len(args) == 0 {
			var body []http2.KVResponse
			var errBody http2.KVResponse

			resp, err := hyprClient.Get("/hypr/kv/", &body, &errBody)
			if err != nil {
				return err
			}

			if resp.IsError() {
				return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), errBody.Error)
			}

			if jsonOutput {
				return printJSON(cmd, body)
			}

			for _, kv := range body {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", kv.Key, kv.Value)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
