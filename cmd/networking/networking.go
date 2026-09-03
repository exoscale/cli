package networking

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

// NetworkingCmd is the root command for networking services.
var NetworkingCmd = &cobra.Command{
	Use:        "networking",
	Short:      "Networking services management",
	Aliases:    []string{"net"},
	SuggestFor: []string{"network", "vpc"},
}

func init() {
	exocmd.RootCmd.AddCommand(NetworkingCmd)
}
