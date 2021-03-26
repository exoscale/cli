package cmd

import (
	"context"
	"fmt"
	"time"

	v2 "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

var sksCmd = &cobra.Command{
	Use:   "sks",
	Short: "Scalable Kubernetes Service management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		// Some SKS operations can take a long time, raising
		// the Exoscale API client timeout as a precaution.
		cs.Client.SetTimeout(10 * time.Minute)
	},
}

func init() {
	RootCmd.AddCommand(sksCmd)
}

// lookupSKSCluster attempts to look up a SKS cluster resource by name or ID.
func lookupSKSCluster(ctx context.Context, zone, v string) (*v2.SKSCluster, error) {
	clusters, err := cs.ListSKSClusters(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("unable to list SKS clusters in zone %s: %v", zone, err)
	}

	for _, cluster := range clusters {
		if cluster.ID == v || cluster.Name == v {
			return cluster, nil
		}
	}

	return nil, fmt.Errorf("SKS cluster %q not found", v) // nolint:golint
}
