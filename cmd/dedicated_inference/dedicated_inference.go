package dedicated_inference

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var DedicatedInferenceCmd = &cobra.Command{
	Use:     "dedicated-inference",
	Short:   "Dedicated AI inference management",
	Aliases: []string{"ai", "inference"},
}

func init() {
	exocmd.RootCmd.AddCommand(DedicatedInferenceCmd)
}
