package iam

import (
	"github.com/spf13/cobra"
)

var iamOrgPolicyCmd = &cobra.Command{
	Use:   "org-policy",
	Short: "IAM Organization Policy management",
}

func init() {
	iamCmd.AddCommand(iamOrgPolicyCmd)
}
