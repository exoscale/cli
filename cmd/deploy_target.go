package cmd

import (
	"context"
	"fmt"

	exov2 "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

var deployTargetCmd = &cobra.Command{
	Use:     "deploytarget",
	Short:   "Deploy Targets management",
	Aliases: []string{"dt"},
}

func init() {
	vmCmd.AddCommand(deployTargetCmd)
}

// lookupDeployTarget attempts to look up a Deploy Target resource by name or ID.
func lookupDeployTarget(ctx context.Context, zone, v string) (*exov2.DeployTarget, error) {
	deployTargets, err := cs.ListDeployTargets(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("unable to list Deploy Targets in zone %s: %v", zone, err)
	}

	for _, dt := range deployTargets {
		if dt.ID == v || dt.Name == v {
			return dt, nil
		}
	}

	return nil, fmt.Errorf("Deploy Target %q not found", v) // nolint:golint
}
