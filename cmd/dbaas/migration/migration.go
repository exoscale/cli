package migration

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:     "migration",
	Short:   "database migration management",
	Aliases: []string{"c"},
}
