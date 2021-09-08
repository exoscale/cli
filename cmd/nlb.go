package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var nlbCmd = &cobra.Command{
	Use:     "load-balancer",
	Short:   "Network Load Balancers management",
	Aliases: []string{"nlb"},
}

var deprecatedNLBCmd = &cobra.Command{
	Use:   "nlb",
	Short: "Network Load Balancers management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo nlb" commands are deprecated and will be removed in a future
version, please use "exo compute load-balancer" replacement commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

func init() {
	computeCmd.AddCommand(nlbCmd)

	// FIXME: remove this someday.
	RootCmd.AddCommand(deprecatedNLBCmd)
}
