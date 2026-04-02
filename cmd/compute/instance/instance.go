package instance

import (
	"errors"
	"fmt"

	"github.com/exoscale/cli/cmd/compute"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

var instanceTypeFamilies = []string{
	"cpu",
	"gpu",
	"gpu2",
	"memory",
	"standard",
	"storage",
}

var instanceTypeSizes = []string{
	"micro",
	"tiny",
	"small",
	"medium",
	"large",
	"extra-large",
	"huge",
	"jumbo",
	"mega",
	"titan",
}

var instanceCmd = &cobra.Command{
	Use:     "instance",
	Short:   "Compute instances management",
	Aliases: []string{"i"},
}

func init() {
	compute.ComputeCmd.AddCommand(instanceCmd)
}

// findInstance looks up an instance by name or ID from a ListInstancesResponse
// and enriches the "not found" error with the zone that was searched,
// reminding the user to use -z to target a different zone.
func findInstance(resp *v3.ListInstancesResponse, nameOrID, zone string) (v3.ListInstancesResponseInstances, error) {
	instance, err := resp.FindListInstancesResponseInstances(nameOrID)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return v3.ListInstancesResponseInstances{}, fmt.Errorf(
				"instance %q not found in zone %s\nHint: use -z <zone> to specify a different zone, or run 'exo compute instance list' to see instances across all zones",
				nameOrID,
				zone,
			)
		}
		return v3.ListInstancesResponseInstances{}, err
	}
	return instance, nil
}
