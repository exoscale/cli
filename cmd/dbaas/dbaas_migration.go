package dbaas

import "github.com/spf13/cobra"

var dbaasMigrationCmd = &cobra.Command{
	Use:     "migration",
	Short:   "database migration management",
	Aliases: []string{"c"},
}

func init() {
	dbaasCmd.AddCommand(dbaasMigrationCmd)
}
