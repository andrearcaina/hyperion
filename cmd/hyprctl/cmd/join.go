package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

type joinResponse struct {
	Status string `json:"status,omitempty"`
}

var (
	joinNodeID   string
	joinNodeAddr string
	joinHTTPAddr string
	joinGRPCAddr string
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a node to the cluster",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := hyprClient.Join(cmd.Context(), joinNodeID, joinNodeAddr, joinHTTPAddr, joinGRPCAddr); err != nil {
			return err
		}

		if jsonOutput {
			return printJSON(cmd, joinResponse{Status: "joined"})
		}

		fmt.Fprintln(cmd.OutOrStdout(), "joined")
		return nil
	},
}

func init() {
	joinCmd.Flags().StringVarP(&joinNodeID, "node-id", "n", "", "Node ID to join")
	joinCmd.Flags().StringVarP(&joinNodeAddr, "node-addr", "A", "", "Node address to join")
	joinCmd.Flags().StringVar(&joinHTTPAddr, "http-addr", "", "Node HTTP address advertised for forwarding")
	joinCmd.Flags().StringVar(&joinGRPCAddr, "grpc-addr", "", "Node gRPC address advertised for forwarding")

	_ = joinCmd.MarkFlagRequired("node-id")
	_ = joinCmd.MarkFlagRequired("node-addr")

	rootCmd.AddCommand(joinCmd)
}
