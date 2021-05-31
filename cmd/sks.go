package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var sksCmd = &cobra.Command{
	Use:   "sks",
	Short: "Scalable Kubernetes Service management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		// Some SKS operations can take a long time, raising
		// the Exoscale API client timeout as a precaution.
		cs.Client.SetTimeout(10 * time.Minute)
	},
}

func init() {
	RootCmd.AddCommand(sksCmd)
}
