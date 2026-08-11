package dbaas

import (
	"github.com/spf13/cobra"
)

var dbaasAclCmd = &cobra.Command{
	Use:   "acl",
	Short: "Manage DBaaS ACL configuration",
}

func init() {
	dbaasCmd.AddCommand(dbaasAclCmd)
}