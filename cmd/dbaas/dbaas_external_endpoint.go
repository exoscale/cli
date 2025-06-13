package dbaas

import (
	"github.com/spf13/cobra"
)

var dbaasExternalEndpointCmd = &cobra.Command{
	Use:   "external-endpoint",
	Short: "Manage DBaaS external endpoints",
}

func init() {
	dbaasCmd.AddCommand(dbaasExternalEndpointCmd)
}
