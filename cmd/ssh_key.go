package cmd

import (
	"github.com/spf13/cobra"
)

var computeSSHKeyCmd = &cobra.Command{
	Use:   "ssh-key",
	Short: "Compute SSH keys management",
}

func init() {
	computeCmd.AddCommand(computeSSHKeyCmd)
}
