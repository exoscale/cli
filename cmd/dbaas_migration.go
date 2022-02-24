package cmd

import "github.com/spf13/cobra"

var dbaasMigrationCmd = &cobra.Command{
	Use:     "migration",
	Short:   "migration status/check",
	Aliases: []string{"c"},
}

func init() {
	dbaasCmd.AddCommand(dbaasMigrationCmd)
}
