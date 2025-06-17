package dns

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS cmd lets you host your zones and manage records",
}

func init() {
	exocmd.RootCmd.AddCommand(dnsCmd)
}
