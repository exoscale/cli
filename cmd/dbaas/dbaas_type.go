package dbaas

import (
	"github.com/spf13/cobra"
)

var dbaasTypeCmd = &cobra.Command{
	Use:   "type",
	Short: "Database Services types management",
}

func init() {
	dbaasCmd.AddCommand(dbaasTypeCmd)
}
