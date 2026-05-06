/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/url"

	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
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
		var errBody http2.KVResponse

		resp, err := hyprClient.Delete("/hypr/kv/"+url.PathEscape(key), &errBody)
		if err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
		if resp.IsError() {
			if errBody.Error != "" {
				return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), errBody.Error)
			}
			return fmt.Errorf("request failed with status %d", resp.StatusCode())
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
