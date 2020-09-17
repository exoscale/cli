package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var instancePoolCmd = &cobra.Command{
	Use:   "instancepool",
	Short: "Instance Pools management",
}

func init() {
	RootCmd.AddCommand(instancePoolCmd)
}

func getInstancePoolByID(id, zone *egoscale.UUID) (*egoscale.InstancePool, error) {
	resp, err := cs.RequestWithContext(gContext, egoscale.GetInstancePool{
		ID:     id,
		ZoneID: zone,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*egoscale.GetInstancePoolResponse)

	return &r.InstancePools[0], nil
}

func getInstancePoolByNameOrID(v string, zone *egoscale.UUID) (*egoscale.InstancePool, error) {
	instancePools := make([]egoscale.InstancePool, 0)

	id, err := egoscale.ParseUUID(v)
	if err == nil {
		return getInstancePoolByID(id, zone)
	}

	resp, err := cs.RequestWithContext(gContext, egoscale.ListInstancePools{
		ZoneID: zone,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*egoscale.ListInstancePoolsResponse)

	for _, i := range r.InstancePools {
		if i.Name == v {
			instancePools = append(instancePools, i)
		}
	}

	switch count := len(instancePools); {
	case count == 0:
		return nil, fmt.Errorf("not found: %q", v)
	case count > 1:
		return nil, fmt.Errorf(`more than one element found: %d`, count)
	}

	return &instancePools[0], nil
}
