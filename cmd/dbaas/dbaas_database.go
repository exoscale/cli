package cmd

import (
	"github.com/spf13/cobra"
)

var dbaasDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Manage DBaaS databases",
}

func init() {
	dbaasCmd.AddCommand(dbaasDatabaseCmd)
}
