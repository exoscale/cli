package iam

import (
	"github.com/spf13/cobra"
)

var iamAPIKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "API Key management",
}

func init() {
	iamCmd.AddCommand(iamAPIKeyCmd)
}
