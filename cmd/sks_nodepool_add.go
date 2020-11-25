package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var sksNodepoolAddCmd = &cobra.Command{
	Use:   "add <cluster name | ID> <Nodepool name>",
	Short: "Add a Nodepool to a SKS cluster",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{
			"instance-type",
			"zone",
		})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			c    = args[0]
			name = args[1]
		)

		z, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone, err := getZoneByNameOrID(z)
		if err != nil {
			return fmt.Errorf("error retrieving zone: %s", err)
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		size, err := cmd.Flags().GetInt64("size")
		if err != nil {
			return err
		}

		instanceType, err := cmd.Flags().GetString("instance-type")
		if err != nil {
			return err
		}
		serviceOffering, err := getServiceOfferingByNameOrID(instanceType)
		if err != nil {
			return fmt.Errorf("error retrieving service offering: %s", err)
		}

		diskSize, err := cmd.Flags().GetInt64("disk-size")
		if err != nil {
			return err
		}

		securityGroups, err := cmd.Flags().GetStringSlice("security-group")
		if err != nil {
			return err
		}

		var securityGroupIDs []egoscale.UUID
		if len(securityGroupIDs) > 0 {
			securityGroupIDs, err = getSecurityGroupIDs(securityGroups)
			if err != nil {
				return err
			}
		}

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone.Name))
		cluster, err := lookupSKSCluster(ctx, zone.Name, c)
		if err != nil {
			return err
		}

		nodepool, err := cluster.AddNodepool(ctx, &egoscale.SKSNodepool{
			Name:           name,
			Description:    description,
			Size:           size,
			InstanceTypeID: serviceOffering.ID.String(),
			DiskSize:       diskSize,
			SecurityGroupIDs: func() []string {
				sgs := make([]string, len(securityGroupIDs))
				for i := range securityGroupIDs {
					sgs[i] = securityGroupIDs[i].String()
				}
				return sgs
			}(),
		})
		if err != nil {
			return fmt.Errorf("error adding Nodepool to the SKS cluster: %s", err)
		}

		if !gQuiet {
			return output(showSKSNodepool(zone, cluster.ID, nodepool.ID))
		}

		return nil
	},
}

func init() {
	sksNodepoolAddCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolAddCmd.Flags().String("description", "", "description")
	sksNodepoolAddCmd.Flags().Int64("size", 2, "Nodepool size")
	sksNodepoolAddCmd.Flags().String("instance-type", defaultServiceOffering,
		"Nodepool Compute instances type")
	sksNodepoolAddCmd.Flags().Int64("disk-size", 50,
		"Nodepool Compute instances disk size")
	sksNodepoolAddCmd.Flags().StringSlice("security-group", nil,
		"Nodepool Security Group <name | id> (can be specified multiple times)")
	sksNodepoolCmd.AddCommand(sksNodepoolAddCmd)
}
