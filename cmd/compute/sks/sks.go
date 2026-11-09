package sks

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/cmd/compute"
)

const defaultInstanceType = "standard.medium"

var sksCmd = &cobra.Command{
	Use:   "sks",
	Short: "Scalable Kubernetes Service management",
}

func init() {
	compute.ComputeCmd.AddCommand(sksCmd)
}
