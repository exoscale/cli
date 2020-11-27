package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var sksNodepoolScaleCmd = &cobra.Command{
	Use:   "scale <cluster name | ID> <Nodepool name | ID> <size>",
	Short: "Scale a SKS cluster Nodepool size",
	Long: `This command scales a SKS cluster Nodepool size up (growing) or down
(shrinking).

In case of a scale-down, operators should use the "exo sks nodepool evict"
variant, allowing them to specify which specific Nodes should be evicted from
the pool rather than leaving the decision to the SKS manager.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c  = args[0]
			np = args[1]

			nodepool *egoscale.SKSNodepool
		)

		size, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid size %q", args[2])
		}
		if size <= 0 {
			return errors.New("minimum Nodepool size is 1")
		}

		z, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone, err := getZoneByNameOrID(z)
		if err != nil {
			return fmt.Errorf("error retrieving zone: %s", err)
		}

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone.Name))
		cluster, err := lookupSKSCluster(ctx, zone.Name, c)
		if err != nil {
			return err
		}

		for _, n := range cluster.Nodepools {
			if n.ID == np || n.Name == np {
				nodepool = n
				break
			}
		}
		if nodepool == nil {
			return errors.New("Nodepool not found") // nolint:golint
		}

		decorateAsyncOperation(fmt.Sprintf("Scaling Nodepool %q...", np), func() {
			err = cluster.ScaleNodepool(ctx, nodepool, int64(size))
		})
		if err != nil {
			return err
		}

		if !gQuiet {
			return output(showSKSNodepool(zone, cluster.ID, nodepool.ID))
		}

		return nil
	},
}

func init() {
	sksNodepoolScaleCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolCmd.AddCommand(sksNodepoolScaleCmd)
}
