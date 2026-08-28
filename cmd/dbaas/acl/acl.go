package acl

import (
	"github.com/spf13/cobra"
)

// Cmd is the root command for DBaaS ACL subcommands.
var Cmd = &cobra.Command{
	Use:   "acl",
	Short: "Manage DBaaS ACL configuration",
}
