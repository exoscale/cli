package deployment

import (
	"github.com/spf13/cobra"
)

// Cmd is the root command for deployment subcommands.
var Cmd = &cobra.Command{
	Use:   "deployment",
	Short: "Manage AI deployments",
}
