package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var instanceTypeFamilies = []string{
	"cpu",
	"gpu",
	"gpu2",
	"memory",
	"standard",
	"storage",
}

var instanceTypeSizes = []string{
	"micro",
	"tiny",
	"small",
	"medium",
	"large",
	"extra-large",
	"huge",
	"jumbo",
	"mega",
	"titan",
}

var instanceCmd = &cobra.Command{
	Use:     "instance",
	Short:   "Compute instances management",
	Aliases: []string{"i"},
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		// Some instance operations can take a long time, raising
		// the Exoscale API client timeout as a precaution.
		cs.Client.SetTimeout(10 * time.Minute)
	},
}

func init() {
	computeCmd.AddCommand(instanceCmd)
}
