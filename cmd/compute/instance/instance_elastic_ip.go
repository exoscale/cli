package instance

import (
	"github.com/spf13/cobra"
)

var instanceEIPCmd = &cobra.Command{
	Use:     "elastic-ip",
	Short:   "Manage Compute instance Elastic IP addresses",
	Aliases: []string{"eip"},
}

func init() {
	instanceCmd.AddCommand(instanceEIPCmd)
}
