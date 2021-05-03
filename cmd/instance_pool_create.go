package cmd

import (
	"fmt"
	"strings"

	exov2 "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var instancePoolCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create an Instance Pool",
	Long: fmt.Sprintf(`This command creates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolShowOutput{}), ", ")),
	Aliases: gCreateAlias,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		instancePool := new(exov2.InstancePool)
		instancePool.Name = args[0]

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		zoneV1, err := getZoneByNameOrID(zone)
		if err != nil {
			return err
		}

		antiAffinityGroups, err := cmd.Flags().GetStringSlice("anti-affinity-group")
		if err != nil {
			return err
		}
		for _, v := range antiAffinityGroups {
			antiAffinityGroup, err := getAntiAffinityGroupByNameOrID(v)
			if err != nil {
				return err
			}
			instancePool.AntiAffinityGroupIDs = append(instancePool.AntiAffinityGroupIDs, antiAffinityGroup.ID.String())
		}

		deployTargetFlagVal, err := cmd.Flags().GetString("deploy-target")
		if err != nil {
			return err
		}
		if deployTargetFlagVal != "" {
			deployTarget, err := lookupDeployTarget(ctx, zone, deployTargetFlagVal)
			if err != nil {
				return err
			}
			instancePool.DeployTargetID = deployTarget.ID
		}

		if instancePool.Description, err = cmd.Flags().GetString("description"); err != nil {
			return err
		}

		if instancePool.DiskSize, err = cmd.Flags().GetInt64("disk"); err != nil {
			return err
		}

		elasticIPs, err := cmd.Flags().GetStringSlice("elastic-ip")
		if err != nil {
			return err
		}
		for _, v := range elasticIPs {
			elasticIP, err := getElasticIPByAddressOrID(v)
			if err != nil {
				return err
			}
			instancePool.ElasticIPIDs = append(instancePool.ElasticIPIDs, elasticIP.ID.String())
		}

		if instancePool.InstancePrefix, err = cmd.Flags().GetString("instance-prefix"); err != nil {
			return err
		}

		if instancePool.IPv6Enabled, err = cmd.Flags().GetBool("ipv6"); err != nil {
			return err
		}

		privateNetworks, err := cmd.Flags().GetStringSlice("privnet")
		if err != nil {
			return err
		}
		for _, v := range privateNetworks {
			privateNetwork, err := getNetwork(v, zoneV1.ID)
			if err != nil {
				return err
			}
			instancePool.PrivateNetworkIDs = append(instancePool.PrivateNetworkIDs, privateNetwork.ID.String())
		}

		securityGroups, err := cmd.Flags().GetStringSlice("security-group")
		if err != nil {
			return err
		}
		for _, v := range securityGroups {
			securityGroup, err := getSecurityGroupByNameOrID(v)
			if err != nil {
				return err
			}
			instancePool.SecurityGroupIDs = append(instancePool.SecurityGroupIDs, securityGroup.ID.String())
		}

		serviceOfferingFlagVal, err := cmd.Flags().GetString("service-offering")
		if err != nil {
			return err
		}
		serviceOffering, err := getServiceOfferingByNameOrID(serviceOfferingFlagVal)
		if err != nil {
			return err
		}
		instancePool.InstanceTypeID = serviceOffering.ID.String()

		if instancePool.Size, err = cmd.Flags().GetInt64("size"); err != nil {
			return err
		}

		if instancePool.SSHKey, err = cmd.Flags().GetString("keypair"); err != nil {
			return err
		}
		if instancePool.SSHKey == "" {
			instancePool.SSHKey = gCurrentAccount.DefaultSSHKey
		}

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

		decorateAsyncOperation(fmt.Sprintf("Creating Instance Pool %q...", instancePool.Name), func() {
			instancePool, err = cs.CreateInstancePool(ctx, zone, instancePool)
		})
		if err != nil {
			return fmt.Errorf("unable to create Instance Pool: %s", err)
		}

		if !gQuiet {
			return output(showInstancePool(zone, instancePool.ID))
		}

		return nil
	},
}

func init() {
	instancePoolCreateCmd.Flags().StringSliceP("anti-affinity-group", "a", nil, "Anti-Affinity Group NAME|ID. Can be specified multiple times.")
	instancePoolCreateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init user data configuration file path")
	instancePoolCreateCmd.Flags().StringP("description", "", "", "Instance Pool description")
	instancePoolCreateCmd.Flags().String("deploy-target", "", "Deploy Target NAME|ID")
	instancePoolCreateCmd.Flags().Int64P("disk", "d", 50, "Instance Pool members disk size")
	instancePoolCreateCmd.Flags().StringSliceP("elastic-ip", "e", nil, "Elastic IP ADDRESS. Can be specified multiple times.")
	instancePoolCreateCmd.Flags().String("instance-prefix", "", "string to prefix Instance Pool member names with")
	instancePoolCreateCmd.Flags().BoolP("ipv6", "6", false, "enable IPv6")
	instancePoolCreateCmd.Flags().StringP("keypair", "k", "", "Instance Pool members SSH key")
	instancePoolCreateCmd.Flags().StringSliceP("privnet", "p", nil, "Private Network NAME|ID. Can be specified multiple times.")
	instancePoolCreateCmd.Flags().StringSliceP("security-group", "s", nil, "Security Group NAME|ID. Can be specified multiple times.")
	instancePoolCreateCmd.Flags().StringP("service-offering", "o", defaultServiceOffering, serviceOfferingHelp)
	instancePoolCreateCmd.Flags().Int64P("size", "", 1, "number of Compute instances in the Instance Pool")
	instancePoolCreateCmd.Flags().StringP("template", "t", defaultTemplate, "Instance Pool members template NAME|ID")
	instancePoolCreateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolCreateCmd.Flags().StringP("zone", "z", "", "Zone to deploy the Instance Pool to")
	instancePoolCmd.AddCommand(instancePoolCreateCmd)
}
