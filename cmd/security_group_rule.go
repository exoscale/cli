package cmd

import (
	"github.com/spf13/cobra"
)

var securityGroupRuleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Security Group rules management",
}

func init() {
	securityGroupCmd.AddCommand(securityGroupRuleCmd)
}
