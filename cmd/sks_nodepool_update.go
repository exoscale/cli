package cmd

import (
	"errors"
	"fmt"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksNodepoolUpdateCmd = &cobra.Command{
	Use:   "update <cluster name | ID> <Nodepool name | ID>",
	Short: "Update a SKS cluster Nodepool",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
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

		z, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone, err := getZoneByNameOrID(z)
		if err != nil {
			return fmt.Errorf("error retrieving zone: %s", err)
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone.Name))
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

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("name") {
			nodepool.Name = name
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("description") {
			nodepool.Description = description
		}

		instanceType, err := cmd.Flags().GetString("instance-type")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("instance-type") {
			serviceOffering, err := getServiceOfferingByNameOrID(instanceType)
			if err != nil {
				return fmt.Errorf("error retrieving service offering: %s", err)
			}
			nodepool.InstanceTypeID = serviceOffering.ID.String()
		}

		diskSize, err := cmd.Flags().GetInt64("disk-size")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("disk-size") {
			nodepool.DiskSize = diskSize
		}

		decorateAsyncOperation(fmt.Sprintf("Updating Nodepool %q...", np), func() {
			err = cluster.UpdateNodepool(ctx, nodepool)
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
	sksNodepoolUpdateCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolUpdateCmd.Flags().String("name", "", "name")
	sksNodepoolUpdateCmd.Flags().String("description", "", "description")
	sksNodepoolUpdateCmd.Flags().String("instance-type", "", "Nodepool Compute instances type")
	sksNodepoolUpdateCmd.Flags().Int64("disk-size", 0, "Nodepool Compute instances disk size")
	sksNodepoolCmd.AddCommand(sksNodepoolUpdateCmd)
}
