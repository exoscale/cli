package ssh_key

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var computeSSHKeyCmd = &cobra.Command{
	Use:   "ssh-key",
	Short: "SSH keys management",
}

func init() {
	exocmd.ComputeCmd.AddCommand(computeSSHKeyCmd)
}
