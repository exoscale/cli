package dbaas

import (
	"github.com/spf13/cobra"
)

var dbaasRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage DBaaS roles",
}

func init() {
	dbaasCmd.AddCommand(dbaasRoleCmd)
}
