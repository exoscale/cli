package kms

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var KMSCmd = &cobra.Command{
	Use:   "kms",
	Short: "Key management",
}

func init() {
	exocmd.RootCmd.AddCommand(KMSCmd)
}
