package cmd

import (
	"github.com/spf13/cobra"
)

var labCmd = &cobra.Command{
	Use:   "lab",
	Short: "Experimental commands",
	Long: `These commands provide work-in-progress functionalities that may or
may not be promoted to production someday.

/!\ IMPORTANT: Exoscale provides no guarantees regarding the stability of the
commands provided in this section, and their syntax can change without prior
notice.`,
}

func init() {
	RootCmd.AddCommand(labCmd)
}
