package cmd

import (
	"github.com/spf13/cobra"
)

var dbaasExternalIntegrationSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "External integration settings management",
}

func init() {
	dbaasExternalIntegrationCmd.AddCommand(dbaasExternalIntegrationSettingsCmd)
}
