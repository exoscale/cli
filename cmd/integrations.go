package cmd

import (
	"github.com/spf13/cobra"
)

var integrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "External tooling integrations",
}

func init() {
	RootCmd.AddCommand(integrationsCmd)
}
