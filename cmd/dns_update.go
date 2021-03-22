package cmd

import (
	"github.com/spf13/cobra"
)

var dnsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update domain record",
}

func init() {
	dnsCmd.AddCommand(dnsUpdateCmd)
}
