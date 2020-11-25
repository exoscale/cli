package cmd

import (
	"errors"
	"fmt"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var sksNodepoolEvictCmd = &cobra.Command{
	Use:   "evict <cluster name | ID> <Nodepool name | ID> <ID [...]>",
	Short: "Evict Nodes from a SKS cluster Nodepool",
	Long: `This command shrinks a SKS cluster Nodepool by removing specified members
(Compute Instance ID).`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c       = args[0]
			np      = args[1]
			members = args[2:]

			nodepool *egoscale.SKSNodepool
		)

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

		if err = cluster.EvictNodepoolMembers(ctx, nodepool, members); err != nil {
			return fmt.Errorf("error updating Nodepool: %s", err)
		}

		if !gQuiet {
			return output(showSKSNodepool(zone, cluster.ID, nodepool.ID))
		}

		return nil
	},
}

func init() {
	sksNodepoolEvictCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolEvictCmd.Flags().StringSliceP("evict", "e", nil,
		"Nodepool member (Compute instance ID) to evict. Can be specified multiple times.")
	// sksNodepoolCmd.AddCommand(sksNodepoolEvictCmd) // TODO: enable this command once it's actually supported by the API
}
