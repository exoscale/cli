package cmd

import "github.com/spf13/cobra"

var dbaasMigrationCmd = &cobra.Command{
	Use:     "migration",
	Short:   "control database migration",
	Aliases: []string{"c"},
}

func init() {
	dbaasCmd.AddCommand(dbaasMigrationCmd)
}
