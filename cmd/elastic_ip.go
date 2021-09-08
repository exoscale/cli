package cmd

import (
	"github.com/spf13/cobra"
)

var elasticIPCmd = &cobra.Command{
	Use:     "elastic-ip",
	Short:   "Elastic IP addresses management",
	Aliases: []string{"eip"},
}

func init() {
	computeCmd.AddCommand(elasticIPCmd)
}
