package cmd

import (
	"github.com/spf13/cobra"
)

// runstatusCmd represents the runstatus command
var runstatusCmd = &cobra.Command{
	Use:   "runstatus",
	Short: "Manage your Runstat.us pages",
	Long: `Focus on building your service,
knowing that when something does go wrong you can keep everyone informed using Runstatus.`,
}

func init() {
	RootCmd.AddCommand(runstatusCmd)
}
