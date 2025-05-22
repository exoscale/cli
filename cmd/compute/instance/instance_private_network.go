package instance

import (
	"github.com/spf13/cobra"
)

var instancePrivnetCmd = &cobra.Command{
	Use:     "private-network",
	Short:   "Manage Compute instance Private Networks",
	Aliases: []string{"privnet"},
}

func init() {
	instanceCmd.AddCommand(instancePrivnetCmd)
}
