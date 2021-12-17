package cmd

import (
	"github.com/spf13/cobra"
)

var dbaasIntegrationCmd = &cobra.Command{
	Use:     "integration",
	Aliases: []string{"integ"},
	Short:   "Database Service integrations management",
}

func init() {
	dbaasCmd.AddCommand(dbaasIntegrationCmd)
}
