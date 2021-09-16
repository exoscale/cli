package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var instancePoolCmd = &cobra.Command{
	Use:     "instance-pool",
	Short:   "Instance Pools management",
	Aliases: []string{"pool"},
}

var deprecatedInstancePoolCmd = &cobra.Command{
	Use:   "instancepool",
	Short: "Instance Pools management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo instancepool" commands are deprecated and will be removed in
a future version, please use "exo compute instance-pool" replacement
commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

func init() {
	computeCmd.AddCommand(instancePoolCmd)

	// FIXME: remove this someday.
	RootCmd.AddCommand(deprecatedInstancePoolCmd)
}
