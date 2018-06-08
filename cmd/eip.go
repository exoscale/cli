package cmd

import (
	"github.com/spf13/cobra"
)

// eipCmd represents the eip command
var eipCmd = &cobra.Command{
	Use:   "eip",
	Short: "Elastic IPs management",
}

func init() {
	RootCmd.AddCommand(eipCmd)
}
