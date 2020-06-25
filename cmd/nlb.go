package cmd

import (
	"context"
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var nlbCmd = &cobra.Command{
	Use:   "nlb",
	Short: "Network Load Balancers management",
}

func init() {
	RootCmd.AddCommand(nlbCmd)
}

// lookupNLB attempts to look up an NLB resource by name or ID.
func lookupNLB(ctx context.Context, zone, ref string) (*egoscale.NetworkLoadBalancer, error) {
	nlbs, err := cs.ListNetworkLoadBalancers(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("unable to list Network Load Balancers in zone %s: %v", zone, err)
	}

	for _, nlb := range nlbs {
		if nlb.ID == ref || nlb.Name == ref {
			return nlb, nil
		}
	}

	return nil, fmt.Errorf("Network Load Balancer %q not found", ref) // nolint:golint
}
