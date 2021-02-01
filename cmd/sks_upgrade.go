package cmd

import (
	"fmt"

	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var sksUpgradeCmd = &cobra.Command{
	Use:   "upgrade <name | ID> <version>",
	Short: "Upgrade a SKS cluster Kubernetes version",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c       = args[0]
			version = args[1]
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
		cluster, err := lookupSKSCluster(ctx, zone, c)
		if err != nil {
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Upgrading SKS cluster %q...", c), func() {
			err = cs.UpgradeSKSCluster(ctx, zone, cluster.ID, version)
		})
		if err != nil {
			return err
		}

		if !gQuiet {
			return output(showSKSCluster(zone, cluster.ID))
		}

		return nil
	},
}

func init() {
	sksUpgradeCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksCmd.AddCommand(sksUpgradeCmd)
}