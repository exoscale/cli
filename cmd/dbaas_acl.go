package cmd

import (
	"github.com/spf13/cobra"
)

var dbaasAclCmd = &cobra.Command{
	Use:   "acl",
	Short: "Manage DBaaS acl",
}

func init() {
	dbaasCmd.AddCommand(dbaasAclCmd)
}
