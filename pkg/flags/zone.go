package flags

import "github.com/spf13/cobra"

// Zone flag for list commands
// zone or all zones
func AddZoneListFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("zone", "z", "", "Exoscale zone")
}
