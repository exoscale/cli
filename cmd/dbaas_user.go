package cmd

import (
	"github.com/spf13/cobra"
)

var dbaasUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage DBaaS users",
}

func init() {
	dbaasCmd.AddCommand(dbaasUserCmd)
}
