package cmd

import "github.com/spf13/cobra"

var nlbServiceCmd = &cobra.Command{
	Use:     "service",
	Short:   "Manage Network Load Balancer services",
	Aliases: []string{"svc"},
}

func init() {
	nlbCmd.AddCommand(nlbServiceCmd)
}
