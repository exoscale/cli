package cmd

import (
	"errors"
	"fmt"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksNodepoolResetFields = []string{
	"anti-affinity-groups",
	"deploy-target",
	"description",
	"security-groups",
}

var sksNodepoolUpdateCmd = &cobra.Command{
	Use:   "update CLUSTER-NAME|ID NODEPOOL-NAME|ID",
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
			updated  bool
		)

		zone, err := cmd.Flags().GetString("zone")
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

		resetFields, err := cmd.Flags().GetStringSlice("reset")
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("name") {
			if nodepool.Name, err = cmd.Flags().GetString("name"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("deploy-target") {
			deployTargetFlagVal, err := cmd.Flags().GetString("deploy-target")
			if err != nil {
				return err
			}
			if deployTargetFlagVal != "" {
				deployTarget, err := lookupDeployTarget(ctx, zone, deployTargetFlagVal)
				if err != nil {
					return fmt.Errorf("error retrieving Deploy Target: %s", err)
				}
				nodepool.DeployTargetID = deployTarget.ID
				updated = true
			}
		}

		if cmd.Flags().Changed("description") {
			if nodepool.Description, err = cmd.Flags().GetString("description"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("instance-prefix") {
			if nodepool.InstancePrefix, err = cmd.Flags().GetString("instance-prefix"); err != nil {
				return err
			}
			updated = true
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
			updated = true
		}

		if cmd.Flags().Changed("disk-size") {
			if nodepool.DiskSize, err = cmd.Flags().GetInt64("disk-size"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("anti-affinity-group") {
			antiAffinityGroups, err := cmd.Flags().GetStringSlice("anti-affinity-group")
			if err != nil {
				return err
			}

			antiAffinityGroupIDs, err := getAffinityGroupIDs(antiAffinityGroups)
			if err != nil {
				return err
			}
			nodepool.AntiAffinityGroupIDs = func() []string {
				ids := make([]string, len(antiAffinityGroups))
				for i := range antiAffinityGroupIDs {
					ids[i] = antiAffinityGroupIDs[i].String()
				}
				return ids
			}()
			updated = true
		}

		if cmd.Flags().Changed("security-group") {
			securityGroups, err := cmd.Flags().GetStringSlice("security-group")
			if err != nil {
				return err
			}

			securityGroupIDs, err := getSecurityGroupIDs(securityGroups)
			if err != nil {
				return err
			}
			nodepool.SecurityGroupIDs = func() []string {
				ids := make([]string, len(securityGroups))
				for i := range securityGroupIDs {
					ids[i] = securityGroupIDs[i].String()
				}
				return ids
			}()
			updated = true
		}

		decorateAsyncOperation(fmt.Sprintf("Updating Nodepool %q...", np), func() {
			if updated {
				if err = cluster.UpdateNodepool(ctx, nodepool); err != nil {
					return
				}
			}

			for _, f := range resetFields {
				switch f {
				case "anti-affinity-groups":
					err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.AntiAffinityGroupIDs)
				case "deploy-target":
					err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.DeployTargetID)
				case "description":
					err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.Description)
				case "security-groups":
					err = cluster.ResetNodepoolField(ctx, nodepool, &nodepool.SecurityGroupIDs)
				}
				if err != nil {
					return
				}
			}
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
	sksNodepoolUpdateCmd.Flags().String("name", "", "Nodepool name")
	sksNodepoolUpdateCmd.Flags().String("deploy-target", "", "Nodepool Deploy Target NAME|ID")
	sksNodepoolUpdateCmd.Flags().String("description", "", "Nodepool description")
	sksNodepoolUpdateCmd.Flags().String("instance-prefix", "", "string to prefix Nodepool member names with")
	sksNodepoolUpdateCmd.Flags().String("instance-type", "", "Nodepool Compute instances type")
	sksNodepoolUpdateCmd.Flags().Int64("disk-size", 0, "Nodepool Compute instances disk size")
	sksNodepoolUpdateCmd.Flags().StringSlice("anti-affinity-group", nil,
		"Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times). "+
			"Note: this replaces the current value, it is not cumulative.")
	sksNodepoolUpdateCmd.Flags().StringSlice("security-group", nil,
		"Nodepool Security Group NAME|ID (can be specified multiple times)"+
			"Note: this replaces the current value, it is not cumulative.")
	sksNodepoolUpdateCmd.Flags().StringSliceP("reset", "r", nil, fmt.Sprintf("properties to reset to default value. Supported values: %s", strings.Join(sksNodepoolResetFields, ", ")))
	sksNodepoolCmd.AddCommand(sksNodepoolUpdateCmd)
}
