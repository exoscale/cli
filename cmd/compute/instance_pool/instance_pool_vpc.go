// AI-modified by claude-opus-5-5 - not reviewed yet
package instance_pool

import (
	"context"
	"errors"
	"fmt"

	"github.com/exoscale/cli/cmd/networking/vpc"
	v3 "github.com/exoscale/egoscale/v3"
)

// resolveVPC resolves a VPC NAME|ID and a list of Subnet NAME|IDs to their IDs.
func resolveVPC(
	ctx context.Context,
	client *v3.Client,
	vpcNameOrID string,
	subnets []string,
) (v3.UUID, []v3.UUID, error) {
	vpcEntry, err := vpc.FindVPC(ctx, client, vpcNameOrID)
	if err != nil {
		return "", nil, err
	}

	vpcSubnets, err := client.ListSubnets(ctx, vpcEntry.ID)
	if err != nil {
		return "", nil, fmt.Errorf("error listing VPC Subnets: %w", err)
	}

	subnetIDs := make([]v3.UUID, 0, len(subnets))
	for _, nameOrID := range subnets {
		subnet, err := vpcSubnets.FindListSubnetEntry(nameOrID)
		if err != nil {
			if errors.Is(err, v3.ErrNotFound) {
				return "", nil, fmt.Errorf("subnet %q not found in VPC %q", nameOrID, vpcNameOrID)
			}
			return "", nil, fmt.Errorf("error retrieving VPC Subnet: %w", err)
		}
		subnetIDs = append(subnetIDs, subnet.ID)
	}

	return vpcEntry.ID, subnetIDs, nil
}
