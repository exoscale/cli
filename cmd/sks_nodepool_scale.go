package cmd

import (
	"errors"
	"fmt"
	"strconv"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksNodepoolScaleCmd = &cobra.Command{
	Use:   "scale CLUSTER-NAME|ID NODEPOOL-NAME|ID SIZE",
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

			nodepool *exov2.SKSNodepool
		)

		size, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid size %q", args[2])
		}
		if size <= 0 {
			return errors.New("minimum Nodepool size is 1")
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, c)
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

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to scale Nodepool %q to %d?", nodepool.Name, size)) {
				return nil
			}
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
	sksNodepoolScaleCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	sksNodepoolScaleCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolCmd.AddCommand(sksNodepoolScaleCmd)
}
