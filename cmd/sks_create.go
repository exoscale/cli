package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

var sksCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a SKS clsuter",
	Long: fmt.Sprintf(`This command creates a SKS cluster.

Note: SKS cluster Nodes' kubelet configuration is set to use the Exoscale
Cloud Controller Manager (CCM) as Cloud Provider by default. Cluster Nodes
will remain in the "NotReady" status until the Exoscale CCM is deployed by
cluster operators. Please refer to the Exoscale CCM documentation for more
information:

    https://github.com/exoscale/exoscale-cloud-controller-manager

If you do not want to use a Cloud Controller Manager, add the
"--no-exoscale-ccm" option to the command. This cannot be changed once the
cluster has been created.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksShowOutput{}), ", ")),
	Aliases: gCreateAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{
			"kubernetes-version",
			"zone",
		})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			name    = args[0]
			cluster *egoscale.SKSCluster
		)

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		version, err := cmd.Flags().GetString("kubernetes-version")
		if err != nil {
			return err
		}

		noExoscaleCCM, err := cmd.Flags().GetBool("no-exoscale-ccm")
		if err != nil {
			return err
		}

		ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
		decorateAsyncOperation(fmt.Sprintf("Creating SKS cluster %q...", name), func() {
			cluster, err = cs.CreateSKSCluster(ctx, zone, &egoscale.SKSCluster{
				Name:                           name,
				Description:                    description,
				Version:                        version,
				ExoscaleCloudControllerEnabled: !noExoscaleCCM,
			})
		})
		if err != nil {
			return err
		}

		nodepoolSize, err := cmd.Flags().GetInt64("nodepool-size")
		if err != nil {
			return err
		}

		if nodepoolSize > 0 {
			nodepoolName, err := cmd.Flags().GetString("nodepool-name")
			if err != nil {
				return err
			}
			if nodepoolName == "" {
				nodepoolName = name
			}

			nodepoolDescription, err := cmd.Flags().GetString("nodepool-description")
			if err != nil {
				return err
			}

			nodepoolInstanceType, err := cmd.Flags().GetString("nodepool-instance-type")
			if err != nil {
				return err
			}
			nodepoolServiceOffering, err := getServiceOfferingByNameOrID(nodepoolInstanceType)
			if err != nil {
				return fmt.Errorf("error retrieving service offering: %s", err)
			}

			nodepoolDiskSize, err := cmd.Flags().GetInt64("nodepool-disk-size")
			if err != nil {
				return err
			}

			nodepoolSecurityGroups, err := cmd.Flags().GetStringSlice("nodepool-security-group")
			if err != nil {
				return err
			}

			var nodepoolSecurityGroupIDs []egoscale.UUID
			if len(nodepoolSecurityGroups) > 0 {
				nodepoolSecurityGroupIDs, err = getSecurityGroupIDs(nodepoolSecurityGroups)
				if err != nil {
					return err
				}
			}

			decorateAsyncOperation(fmt.Sprintf("Adding Nodepool %q...", nodepoolName), func() {
				_, err = cluster.AddNodepool(ctx, &egoscale.SKSNodepool{
					Name:           nodepoolName,
					Description:    nodepoolDescription,
					Size:           nodepoolSize,
					InstanceTypeID: nodepoolServiceOffering.ID.String(),
					DiskSize:       nodepoolDiskSize,
					SecurityGroupIDs: func() []string {
						sgs := make([]string, len(nodepoolSecurityGroupIDs))
						for i := range nodepoolSecurityGroupIDs {
							sgs[i] = nodepoolSecurityGroupIDs[i].String()
						}
						return sgs
					}(),
				})
			})
			if err != nil {
				return err
			}
		}

		if !gQuiet {
			return output(showSKSCluster(zone, cluster.ID))
		}

		return nil
	},
}

func init() {
	sksCreateCmd.Flags().StringP("zone", "z", "", "SKS cluster zone")
	sksCreateCmd.Flags().String("description", "", "SKS cluster description")
	sksCreateCmd.Flags().String("kubernetes-version", "1.18.6",
		"SKS cluster control plane Kubernetes version")
	sksCreateCmd.Flags().Bool("no-exoscale-ccm", false,
		"do not deploy the Exoscale Cloud Controller Manager in the Kubernetes control plane")
	sksCreateCmd.Flags().Int64("nodepool-size", 0,
		"default Nodepool size (default: 0). If 0, no default Nodepool will be added to the cluster.")
	sksCreateCmd.Flags().String("nodepool-name", "",
		"default Nodepool name (default: name of the SKS cluster)")
	sksCreateCmd.Flags().String("nodepool-description", "",
		"default Nodepool description")
	sksCreateCmd.Flags().String("nodepool-instance-type", defaultServiceOffering,
		"default Nodepool Compute instances type")
	sksCreateCmd.Flags().Int64("nodepool-disk-size", 50,
		"default Nodepool Compute instances disk size")
	sksCreateCmd.Flags().StringSlice("nodepool-security-group", nil,
		"default Nodepool Security Group <name | id> (can be specified multiple times)")
	sksCmd.AddCommand(sksCreateCmd)
}
