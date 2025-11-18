package sks

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/cmd/compute"
)

var sksCmd = &cobra.Command{
	Use:   "sks",
	Short: "Scalable Kubernetes Service management",
}

func init() {
	compute.ComputeCmd.AddCommand(sksCmd)
}
