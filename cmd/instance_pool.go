package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var instancePoolCmd = &cobra.Command{
	Use:   "instancepool",
	Short: "Instance pool managment",
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
	r := resp.(*egoscale.GetInstancePoolsResponse)

	return &r.ListInstancePoolsResponse[0], nil
}

func getInstancePoolByName(name string, zone *egoscale.UUID) (*egoscale.InstancePool, error) {
	instancePools := []egoscale.InstancePool{}

	id, err := egoscale.ParseUUID(name)
	if err == nil {
		return getInstancePoolByID(id, zone)
	}

	resp, err := cs.RequestWithContext(gContext, egoscale.ListInstancePool{
		ZoneID: zone,
	})
	if err != nil {
		return nil, err
	}
	r := resp.(*egoscale.ListInstancePoolsResponse)

	for _, i := range r.ListInstancePoolsResponse {
		if i.Name == name {
			instancePools = append(instancePools, i)
		}
	}

	switch count := len(instancePools); {
	case count == 0:
		return nil, fmt.Errorf("not found: %q", name)
	case count > 1:
		return nil, fmt.Errorf(`more than one element found: %d`, count)
	}

	return &instancePools[0], nil
}
