package ssh_key

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var computeSSHKeyCmd = &cobra.Command{
	Use:   "ssh-key",
	Short: "SSH keys management",
}

func init() {
	compute.ComputeCmd.AddCommand(computeSSHKeyCmd)
}
