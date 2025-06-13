package iam

import (
	"github.com/spf13/cobra"
)

var iamRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "IAM Role management",
}

func init() {
	iamCmd.AddCommand(iamRoleCmd)
}
