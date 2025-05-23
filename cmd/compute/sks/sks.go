package sks

import (
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
)

var sksCmd = &cobra.Command{
	Use:   "sks",
	Short: "Scalable Kubernetes Service management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		// Some SKS operations can take a long time, raising
		// the Exoscale API client timeout as a precaution.
		globalstate.EgoscaleClient.SetTimeout(10 * time.Minute)
	},
}

func init() {
	exocmd.ComputeCmd.AddCommand(sksCmd)
}
