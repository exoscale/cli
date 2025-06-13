package cmd

import (
	"github.com/spf13/cobra"
)

var dbaasExternalIntegrationCmd = &cobra.Command{
	Use:   "external-integration",
	Short: "Manage DBaaS external integrations",
}

func init() {
	dbaasCmd.AddCommand(dbaasExternalIntegrationCmd)
}
