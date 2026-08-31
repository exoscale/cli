package role

import (
	"github.com/spf13/cobra"
)

// Cmd is the root command for DBaaS role subcommands.
var Cmd = &cobra.Command{
	Use:   "role",
	Short: "Manage DBaaS roles",
}
