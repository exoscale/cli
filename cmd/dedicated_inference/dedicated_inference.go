package dedicated_inference

import (
    exocmd "github.com/exoscale/cli/cmd"
    "github.com/exoscale/cli/cmd/dedicated_inference/deployment"
    "github.com/exoscale/cli/cmd/dedicated_inference/model"
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
