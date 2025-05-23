package cmd

import (
	"github.com/spf13/cobra"
)

var computeSSHKeyCmd = &cobra.Command{
	Use:   "ssh-key",
	Short: "SSH keys management",
}

func init() {
	ComputeCmd.AddCommand(computeSSHKeyCmd)
}
