package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var instancePoolResetFields = []string{
	"anti-affinity-groups",
	"deploy-target",
	"description",
	"elastic-ips",
	"ipv6",
	"private-networks",
	"ssh-key",
	"security-groups",
	"user-data",
}

var instancePoolUpdateCmd = &cobra.Command{
	Use:   "update NAME|ID",
	Short: "Update an Instance Pool",
	Long: fmt.Sprintf(`This command updates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", ")),
	Aliases: gCreateAlias,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		resetFields, err := cmd.Flags().GetStringSlice("reset")
		if err != nil {
			return err
		}
		for _, f := range resetFields {
			if !isInList(instancePoolResetFields, f) {
				cmdExitOnUsageError(cmd, fmt.Sprintf("--reset: unsupported property %q", f))
			}
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var updated bool

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		zoneV1, err := getZoneByNameOrID(zone)
		if err != nil {
			return err
		}

		instancePool, err := lookupInstancePool(ctx, zone, args[0])
		if err != nil {
			return err
		}

		resetFields, err := cmd.Flags().GetStringSlice("reset")
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("anti-affinity-group") {
			antiAffinityGroups, err := cmd.Flags().GetStringSlice("anti-affinity-group")
			if err != nil {
				return err
			}

			instancePool.AntiAffinityGroupIDs = make([]string, 0)
			for _, v := range antiAffinityGroups {
				antiAffinityGroup, err := getAntiAffinityGroupByNameOrID(v)
				if err != nil {
					return err
				}
				instancePool.AntiAffinityGroupIDs = append(
					instancePool.AntiAffinityGroupIDs,
					antiAffinityGroup.ID.String(),
				)
			}
			updated = true
		}

		if cmd.Flags().Changed("deploy-target") {
			deployTargetFlagVal, err := cmd.Flags().GetString("deploy-target")
			if err != nil {
				return err
			}

			deployTarget, err := lookupDeployTarget(ctx, zone, deployTargetFlagVal)
			if err != nil {
				return err
			}
			instancePool.DeployTargetID = deployTarget.ID
			updated = true
		}

		if cmd.Flags().Changed("description") {
			if instancePool.Description, err = cmd.Flags().GetString("description"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("disk") {
			if instancePool.DiskSize, err = cmd.Flags().GetInt64("disk"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("elastic-ip") {
			elasticIPs, err := cmd.Flags().GetStringSlice("elastic-ip")
			if err != nil {
				return err
			}

			instancePool.ElasticIPIDs = make([]string, 0)
			for _, v := range elasticIPs {
				elasticIP, err := getElasticIPByAddressOrID(v)
				if err != nil {
					return err
				}
				instancePool.ElasticIPIDs = append(instancePool.ElasticIPIDs, elasticIP.ID.String())
			}
			updated = true
		}

		if cmd.Flags().Changed("instance-prefix") {
			if instancePool.InstancePrefix, err = cmd.Flags().GetString("instance-prefix"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("ipv6") {
			if instancePool.IPv6Enabled, err = cmd.Flags().GetBool("ipv6"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("name") {
			if instancePool.Name, err = cmd.Flags().GetString("name"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("privnet") {
			privateNetworks, err := cmd.Flags().GetStringSlice("privnet")
			if err != nil {
				return err
			}

			instancePool.PrivateNetworkIDs = make([]string, 0)
			for _, v := range privateNetworks {
				privateNetwork, err := getNetwork(v, zoneV1.ID)
				if err != nil {
					return err
				}
				instancePool.PrivateNetworkIDs = append(instancePool.PrivateNetworkIDs, privateNetwork.ID.String())
			}
			updated = true
		}

		if cmd.Flags().Changed("security-group") {
			securityGroups, err := cmd.Flags().GetStringSlice("security-group")
			if err != nil {
				return err
			}

			instancePool.SecurityGroupIDs = make([]string, 0)
			for _, v := range securityGroups {
				securityGroup, err := getSecurityGroupByNameOrID(v)
				if err != nil {
					return err
				}
				instancePool.SecurityGroupIDs = append(instancePool.SecurityGroupIDs, securityGroup.ID.String())
			}
			updated = true
		}

		if cmd.Flags().Changed("service-offering") {
			serviceOfferingFlagVal, err := cmd.Flags().GetString("service-offering")
			if err != nil {
				return err
			}
			serviceOffering, err := getServiceOfferingByNameOrID(serviceOfferingFlagVal)
			if err != nil {
				return err
			}
			instancePool.InstanceTypeID = serviceOffering.ID.String()
			updated = true
		}

		if cmd.Flags().Changed("keypair") {
			if instancePool.SSHKey, err = cmd.Flags().GetString("keypair"); err != nil {
				return err
			}
			updated = true
		}

		if cmd.Flags().Changed("template") {
			templateFilterFlagVal, err := cmd.Flags().GetString("template-filter")
			if err != nil {
				return err
			}
			templateFilter, err := validateTemplateFilter(templateFilterFlagVal)
			if err != nil {
				return err
			}

			templateFlagVal, err := cmd.Flags().GetString("template")
			if err != nil {
				return err
			}
			template, err := getTemplateByNameOrID(zoneV1.ID, templateFlagVal, templateFilter)
			if err != nil {
				return err
			}
			instancePool.TemplateID = template.ID.String()
			updated = true
		}

		if cmd.Flags().Changed("cloud-init") {
			userData := ""
			userDataPath, err := cmd.Flags().GetString("cloud-init")
			if err != nil {
				return err
			}
			if userDataPath != "" {
				userData, err = getUserDataFromFile(userDataPath)
				if err != nil {
					return err
				}
				instancePool.UserData = userData
			}
			updated = true
		}

		decorateAsyncOperation(fmt.Sprintf("Updating Instance Pool %q...", instancePool.Name), func() {
			if updated {
				if err = cs.UpdateInstancePool(ctx, zone, instancePool); err != nil {
					return
				}
			}

			for _, f := range resetFields {
				switch f {
				case "anti-affinity-groups":
					err = instancePool.ResetField(ctx, &instancePool.AntiAffinityGroupIDs)
				case "elastic-ips":
					err = instancePool.ResetField(ctx, &instancePool.ElasticIPIDs)
				case "deploy-target":
					err = instancePool.ResetField(ctx, &instancePool.DeployTargetID)
				case "description":
					err = instancePool.ResetField(ctx, &instancePool.Description)
				case "ipv6":
					err = instancePool.ResetField(ctx, &instancePool.IPv6Enabled)
				case "private-networks":
					err = instancePool.ResetField(ctx, &instancePool.PrivateNetworkIDs)
				case "security-groups":
					err = instancePool.ResetField(ctx, &instancePool.SecurityGroupIDs)
				case "ssh-key":
					err = instancePool.ResetField(ctx, &instancePool.SSHKey)
				case "user-data":
					err = instancePool.ResetField(ctx, &instancePool.UserData)
				}
				if err != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("size") {
			size, err := cmd.Flags().GetInt64("size")
			if err != nil {
				return err
			}

			if size > 0 {
				fmt.Fprintln(
					os.Stderr,
					`WARNING: the "--size" flag is deprecated and replaced by the `+
						`"exo instancepool scale" command, it will be removed in a future version.`,
				)

				decorateAsyncOperation(fmt.Sprintf("Scaling Instance Pool %q...", instancePool.Name), func() {
					err = instancePool.Scale(ctx, int64(size))
				})
			}
		}

		if !gQuiet {
			return output(showInstancePool(zone, instancePool.ID))
		}

		return nil
	},
}

func init() {
	instancePoolUpdateCmd.Flags().StringSliceP("anti-affinity-group", "a", nil, "Anti-Affinity Group NAME|ID. Can be specified multiple times.")
	instancePoolUpdateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init user data configuration file path")
	instancePoolUpdateCmd.Flags().StringP("description", "", "", "Instance Pool description")
	instancePoolUpdateCmd.Flags().String("deploy-target", "", "Deploy Target NAME|ID")
	instancePoolUpdateCmd.Flags().Int64P("disk", "d", 50, "Instance Pool members disk size")
	instancePoolUpdateCmd.Flags().StringSliceP("elastic-ip", "e", nil, "Elastic IP ADDRESS. Can be specified multiple times.")
	instancePoolUpdateCmd.Flags().String("instance-prefix", "", "string to prefix Instance Pool member names with")
	instancePoolUpdateCmd.Flags().BoolP("ipv6", "6", false, "enable IPv6")
	instancePoolUpdateCmd.Flags().StringP("keypair", "k", "", "Instance Pool members SSH key")
	instancePoolUpdateCmd.Flags().StringP("name", "n", "", "Instance Pool name")
	instancePoolUpdateCmd.Flags().StringSliceP("privnet", "p", nil, "Private Network NAME|ID. Can be specified multiple times.")
	instancePoolUpdateCmd.Flags().StringSliceP("reset", "r", nil, fmt.Sprintf("properties to reset to default value. Supported values: %s", strings.Join(instancePoolResetFields, ", ")))
	instancePoolUpdateCmd.Flags().StringSliceP("security-group", "s", nil, "Security Group NAME|ID. Can be specified multiple times.")
	instancePoolUpdateCmd.Flags().StringP("service-offering", "o", "", serviceOfferingHelp)
	instancePoolUpdateCmd.Flags().Int64P("size", "", 1, "number of Compute instances in the Instance Pool")
	instancePoolUpdateCmd.Flags().StringP("template", "t", "", "Instance Pool members template NAME|ID")
	instancePoolUpdateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolUpdateCmd.Flags().StringP("zone", "z", "", "Zone to deploy the Instance Pool to")
	instancePoolCmd.AddCommand(instancePoolUpdateCmd)
}
