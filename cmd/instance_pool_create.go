package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var instancePoolCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create an Instance Pool",
	Long: fmt.Sprintf(`This command creates an Instance Pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolItemOutput{}), ", ")),
	Aliases: gCreateAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{
			"service-offering",
			"template",
		})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		zoneName, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		zone, err := getZoneByNameOrID(zoneName)
		if err != nil {
			return err
		}

		so, err := cmd.Flags().GetString("service-offering")
		if err != nil {
			return err
		}

		servOffering, err := getServiceOfferingByNameOrID(so)
		if err != nil {
			return err
		}

		templateFilterCmd, err := cmd.Flags().GetString("template-filter")
		if err != nil {
			return err
		}
		templateFilter, err := validateTemplateFilter(templateFilterCmd)
		if err != nil {
			return err
		}

		templateName, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		template, err := getTemplateByNameOrID(zone.ID, templateName, templateFilter)
		if err != nil {
			return err
		}

		keypair, err := cmd.Flags().GetString("keypair")
		if err != nil {
			return err
		}

		if keypair == "" {
			keypair = gCurrentAccount.DefaultSSHKey
		}

		size, err := cmd.Flags().GetInt("size")
		if err != nil {
			return err
		}

		diskSize, err := cmd.Flags().GetInt("disk")
		if err != nil {
			return err
		}

		aag, err := cmd.Flags().GetStringSlice("anti-affinity-group")
		if err != nil {
			return err
		}
		antiAffinityGroups, err := getAffinityGroupIDs(aag)
		if err != nil {
			return err
		}

		sg, err := cmd.Flags().GetStringSlice("security-group")
		if err != nil {
			return err
		}
		securityGroups, err := getSecurityGroupIDs(sg)
		if err != nil {
			return err
		}

		privnet, err := cmd.Flags().GetStringSlice("privnet")
		if err != nil {
			return err
		}

		privnets, err := getPrivnetIDs(privnet, zone.ID)
		if err != nil {
			return err
		}

		ipv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		userDataPath, err := cmd.Flags().GetString("cloud-init")
		if err != nil {
			return err
		}

		userData := ""

		if userDataPath != "" {
			userData, err = getUserDataFromFile(userDataPath)
			if err != nil {
				return err
			}
		}

		r := asyncTasks([]task{
			{
				egoscale.CreateInstancePool{
					Name:                 args[0],
					Description:          description,
					ZoneID:               zone.ID,
					ServiceOfferingID:    servOffering.ID,
					TemplateID:           template.ID,
					KeyPair:              keypair,
					Size:                 size,
					RootDiskSize:         diskSize,
					AntiAffinityGroupIDs: antiAffinityGroups,
					SecurityGroupIDs:     securityGroups,
					NetworkIDs:           privnets,
					IPv6:                 ipv6,
					UserData:             userData,
				},
				fmt.Sprintf("Creating Instance Pool %q", args[0]),
			},
		})
		errs := filterErrors(r)
		if len(errs) > 0 {
			return errs[0]
		}
		pool := r[0].resp.(*egoscale.CreateInstancePoolResponse)

		if !gQuiet {
			return showInstancePool(pool.ID.String(), pool.ZoneID.String())
		}

		return nil
	},
}

func init() {
	// Required Flags
	instancePoolCreateCmd.Flags().StringP("service-offering", "o", "", serviceOfferingHelp)
	instancePoolCreateCmd.Flags().StringP("template", "t", "", "Instance pool template")

	instancePoolCreateCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCreateCmd.Flags().IntP("size", "", 3, "Number of instance in the pool")
	instancePoolCreateCmd.Flags().IntP("disk", "d", 50, "Disk size")
	instancePoolCreateCmd.Flags().StringP("description", "", "", "Instance pool description")
	instancePoolCreateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init file path")
	instancePoolCreateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolCreateCmd.Flags().StringP("keypair", "k", "", "Instance pool keypair")
	instancePoolCreateCmd.Flags().StringSliceP("anti-affinity-group", "a", nil, "Anti-Affinity group NAME|ID|NAME|ID. Can be specified multiple times.")
	instancePoolCreateCmd.Flags().StringSliceP("security-group", "s", nil, "Security Group NAME|ID|NAME|ID. Can be specified multiple times.")
	instancePoolCreateCmd.Flags().StringSliceP("privnet", "p", nil, "Private Network NAME|ID|NAME|ID. Can be specified multiple times.")
	instancePoolCreateCmd.Flags().BoolP("ipv6", "6", false, "Enable IPv6")
	instancePoolCmd.AddCommand(instancePoolCreateCmd)
}
