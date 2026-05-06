package cmd

import (
	"fmt"

	http2 "github.com/andrearcaina/hyperion/internal/transport/http"
	"github.com/spf13/cobra"
)

type joinResponse struct {
	Status string `json:"status,omitempty"`
}

var (
	joinNodeID   string
	joinNodeAddr string
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a node to the cluster",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var respBody joinResponse
		var errBody http2.KVResponse

		resp, err := hyprClient.Post(
			"/hypr/raft/join",
			http2.JoinRequest{
				NodeID:  joinNodeID,
				Address: joinNodeAddr,
			},
			&respBody,
			&errBody,
		)
		if err != nil {
			return err
		}

		if resp.IsError() {
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), errBody.Error)
		}

		status := respBody.Status
		if status == "" {
			status = "joined"
		}

		if jsonOutput {
			return printJSON(cmd, respBody)
		}

		fmt.Fprintln(cmd.OutOrStdout(), status)
		return nil
	},
}

func init() {
	joinCmd.Flags().StringVarP(&joinNodeID, "node-id", "n", "", "Node ID to join")
	joinCmd.Flags().StringVarP(&joinNodeAddr, "node-addr", "A", "", "Node address to join")

	_ = joinCmd.MarkFlagRequired("node-id")
	_ = joinCmd.MarkFlagRequired("node-addr")

	rootCmd.AddCommand(joinCmd)
}
