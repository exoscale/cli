package vpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/cmd/compute"
	v3 "github.com/exoscale/egoscale/v3"
)

// Cmd is the root command for VPC subcommands.
var Cmd = &cobra.Command{
	Use:   "vpc",
	Short: "Virtual Private Cloud management",
}

func init() {
	compute.ComputeCmd.AddCommand(Cmd)
}

// FindVPC resolves a VPC by name or ID in the client's current zone.
func FindVPC(ctx context.Context, client *v3.Client, nameOrID string) (v3.ListVpcEntry, error) {
	resp, err := client.ListVpcs(ctx)
	if err != nil {
		return v3.ListVpcEntry{}, err
	}

	vpc, err := resp.FindListVpcEntry(nameOrID)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return v3.ListVpcEntry{}, fmt.Errorf(
				"vpc %q not found\nHint: use -z <zone> to specify a different zone, or run 'exo compute vpc list' to see VPCs across all zones",
				nameOrID)
		}
		return v3.ListVpcEntry{}, err
	}

	return vpc, nil
}

// FindSubnet resolves a Subnet by name or ID within the given VPC.
func FindSubnet(ctx context.Context, client *v3.Client, vpcID v3.UUID, nameOrID string) (v3.ListSubnetEntry, error) {
	resp, err := client.ListSubnets(ctx, vpcID)
	if err != nil {
		return v3.ListSubnetEntry{}, err
	}

	subnet, err := resp.FindListSubnetEntry(nameOrID)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return v3.ListSubnetEntry{}, fmt.Errorf("subnet %q not found in VPC %s", nameOrID, vpcID)
		}
		return v3.ListSubnetEntry{}, err
	}

	return subnet, nil
}
