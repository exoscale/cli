package cmd

import (
	"github.com/spf13/cobra"
)

var dbTypeCmd = &cobra.Command{
	Use:   "type",
	Short: "Database Services types management",
}

func init() {
	dbaasCmd.AddCommand(dbTypeCmd)
}
