package cmd

import (
	"context"
	"fmt"

	exov2 "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

var instancePoolCmd = &cobra.Command{
	Use:   "instancepool",
	Short: "Instance Pools management",
}

func init() {
	RootCmd.AddCommand(instancePoolCmd)
}

// lookupInstancePool attempts to look up an Instance Pool resource by name or ID.
func lookupInstancePool(ctx context.Context, zone, v string) (*exov2.InstancePool, error) {
	instancePools, err := cs.ListInstancePools(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("unable to list Instance Pools in zone %s: %v", zone, err)
	}

	for _, instancePool := range instancePools {
		if instancePool.ID == v || instancePool.Name == v {
			return instancePool, nil
		}
	}

	return nil, fmt.Errorf("Instance Pool %q not found", v) // nolint:golint
}
