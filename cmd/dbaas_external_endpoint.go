package cmd

import (
	"github.com/spf13/cobra"
)

var dbaasExternalEndpointCmd = &cobra.Command{
	Use:   "external-endpoint",
	Short: "Database as a Services External Endpoint management",
}

func init() {
	dbaasCmd.AddCommand(dbaasExternalEndpointCmd)
}
