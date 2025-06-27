package elastic_ip

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var elasticIPCmd = &cobra.Command{
	Use:     "elastic-ip",
	Short:   "Elastic IP addresses management",
	Aliases: []string{"eip"},
}

func init() {
	compute.ComputeCmd.AddCommand(elasticIPCmd)
}
