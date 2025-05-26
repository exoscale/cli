package elastic_ip

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var elasticIPCmd = &cobra.Command{
	Use:     "elastic-ip",
	Short:   "Elastic IP addresses management",
	Aliases: []string{"eip"},
}

func init() {
	exocmd.ComputeCmd.AddCommand(elasticIPCmd)
}
