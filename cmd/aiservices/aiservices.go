package aiservices

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/aiservices/deployment"
	"github.com/exoscale/cli/cmd/aiservices/model"
	"github.com/spf13/cobra"
)

var DedicatedInferenceCmd = &cobra.Command{
	Use:     "dedicated-inference",
	Short:   "Dedicated AI inference management",
	Aliases: []string{"ai", "inference"},
}

func init() {
	exocmd.RootCmd.AddCommand(DedicatedInferenceCmd)
	// Attach subcommand groups
	DedicatedInferenceCmd.AddCommand(model.Cmd)
	DedicatedInferenceCmd.AddCommand(deployment.Cmd)
}
