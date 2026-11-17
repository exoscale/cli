package dns

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/dns/add"
	"github.com/exoscale/cli/cmd/dns/update"
)

var Cmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS cmd lets you host your zones and manage records",
}

func init() {
	exocmd.RootCmd.AddCommand(Cmd)
	// Attach subcommand groups that live in their own packages.
	Cmd.AddCommand(add.Cmd)
	Cmd.AddCommand(update.Cmd)
}
