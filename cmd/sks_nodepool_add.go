package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var sksNodepoolAddCmd = &cobra.Command{
	Use:   "add CLUSTER-NAME|ID NODEPOOL-NAME",
	Short: "Add a Nodepool to a SKS cluster",
	Long: fmt.Sprintf(`This command adds a Nodepool to a SKS cluster.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksNodepoolShowOutput{}), ", ")),
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
			c        = args[0]
			name     = args[1]
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

		antiAffinityGroups, err := cmd.Flags().GetStringSlice("anti-affinity-group")
		if err != nil {
			return err
		}

		var antiAffinityGroupIDs []egoscale.UUID
		if len(antiAffinityGroups) > 0 {
			antiAffinityGroupIDs, err = getAffinityGroupIDs(antiAffinityGroups)
			if err != nil {
				return err
			}
		}

		securityGroups, err := cmd.Flags().GetStringSlice("security-group")
		if err != nil {
			return err
		}

		var securityGroupIDs []egoscale.UUID
		if len(securityGroups) > 0 {
			securityGroupIDs, err = getSecurityGroupIDs(securityGroups)
			if err != nil {
				return err
			}
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone.Name))
		cluster, err := lookupSKSCluster(ctx, zone.Name, c)
		if err != nil {
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", name), func() {
			nodepool, err = cluster.AddNodepool(ctx, &exov2.SKSNodepool{
				Name:           name,
				Description:    description,
				Size:           size,
				InstanceTypeID: serviceOffering.ID.String(),
				DiskSize:       diskSize,
				AntiAffinityGroupIDs: func() []string {
					aags := make([]string, len(antiAffinityGroups))
					for i := range antiAffinityGroupIDs {
						aags[i] = antiAffinityGroupIDs[i].String()
					}
					return aags
				}(),
				SecurityGroupIDs: func() []string {
					sgs := make([]string, len(securityGroupIDs))
					for i := range securityGroupIDs {
						sgs[i] = securityGroupIDs[i].String()
					}
					return sgs
				}(),
			})
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
	sksNodepoolAddCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksNodepoolAddCmd.Flags().String("description", "", "description")
	sksNodepoolAddCmd.Flags().Int64("size", 2, "Nodepool size")
	sksNodepoolAddCmd.Flags().String("instance-type", defaultServiceOffering,
		"Nodepool Compute instances type")
	sksNodepoolAddCmd.Flags().Int64("disk-size", 50,
		"Nodepool Compute instances disk size")
	sksNodepoolAddCmd.Flags().StringSlice("anti-affinity-group", nil,
		"Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)")
	sksNodepoolAddCmd.Flags().StringSlice("security-group", nil,
		"Nodepool Security Group NAME|ID (can be specified multiple times)")
	sksNodepoolCmd.AddCommand(sksNodepoolAddCmd)
}
