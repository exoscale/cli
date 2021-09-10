package cmd

import (
	"github.com/spf13/cobra"
)

var securityGroupSourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Security Group external sources management",
}

func init() {
	securityGroupCmd.AddCommand(securityGroupSourceCmd)
}
