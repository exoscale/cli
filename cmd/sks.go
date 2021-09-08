package cmd

import (
	"fmt"
	"os"
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

var deprecatedSKSCmd = &cobra.Command{
	Use:   "sks",
	Short: "Scalable Kubernetes Service management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		// Some SKS operations can take a long time, raising
		// the Exoscale API client timeout as a precaution.
		cs.Client.SetTimeout(10 * time.Minute)

		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo sks" commands are deprecated and will be removed in a future
version, please use "exo compute sks" replacement commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

func init() {
	computeCmd.AddCommand(sksCmd)

	// FIXME: remove this someday.
	RootCmd.AddCommand(deprecatedSKSCmd)
}
