package cmd

import (
	"github.com/spf13/cobra"
)

// labCmd represents the lab command
var labCmd = &cobra.Command{
	Use:   "lab",
	Short: "Experimental commands",
	Long: `These commands provide work-in-progress functionalities that may or
may not be promoted to production someday. Caution: wet paint.`,
}

func init() {
	RootCmd.AddCommand(labCmd)
}
