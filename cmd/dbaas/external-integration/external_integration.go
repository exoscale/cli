package externalintegration

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/cmd/dbaas/external-integration/settings"
)

var Cmd = &cobra.Command{
	Use:   "external-integration",
	Short: "Manage DBaaS external integrations",
}

func init() {
	Cmd.AddCommand(settings.Cmd)
}
