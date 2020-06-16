package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var instancePoolCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an instance pool",
	Long: fmt.Sprintf(`This command creates an instance pool.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolItemOutput{}), ", ")),
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		zoneName, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zoneName == "" {
			zoneName = gCurrentAccount.DefaultZone
		}

		zone, err := getZoneByName(zoneName)
		if err != nil {
			return err
		}

		so, err := cmd.Flags().GetString("service-offering")
		if err != nil {
			return err
		}

		servOffering, err := getServiceOfferingByName(so)
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

		template, err := getTemplateByName(zone.ID, templateName, templateFilter)
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

		sg, err := cmd.Flags().GetStringSlice("security-group")
		if err != nil {
			return err
		}

		securityGroups, err := getSecurityGroups(sg)
		if err != nil {
			return err
		}

		privnet, err := cmd.Flags().GetStringSlice("privnet")
		if err != nil {
			return err
		}

		privnets, err := getPrivnetList(privnet, zone.ID)
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
					Name:              args[0],
					Description:       description,
					ZoneID:            zone.ID,
					ServiceOfferingID: servOffering.ID,
					TemplateID:        template.ID,
					KeyPair:           keypair,
					Size:              size,
					RootDiskSize:      diskSize,
					SecurityGroupIDs:  securityGroups,
					NetworkIDs:        privnets,
					UserData:          userData,
				},
				fmt.Sprintf("Creating instance pool %q", args[0]),
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
	if err := instancePoolCreateCmd.MarkFlagRequired("service-offering"); err != nil {
		log.Fatal(err)
	}

	instancePoolCreateCmd.Flags().StringP("template", "t", "", "Instance pool template")
	if err := instancePoolCreateCmd.MarkFlagRequired("template"); err != nil {
		log.Fatal(err)
	}

	instancePoolCreateCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCreateCmd.Flags().IntP("size", "", 3, "Number of instance in the pool")
	instancePoolCreateCmd.Flags().IntP("disk", "d", 50, "Disk size")
	instancePoolCreateCmd.Flags().StringP("description", "", "", "Instance pool description")
	instancePoolCreateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init file path")
	instancePoolCreateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolCreateCmd.Flags().StringP("keypair", "k", "", "Instance pool keypair")
	instancePoolCreateCmd.Flags().StringSliceP("security-group", "s", nil, "Security groups <name | id, name | id, ...>")
	instancePoolCreateCmd.Flags().StringSliceP("privnet", "p", nil, "Privnets <name | id, name | id, ...>")
	instancePoolCmd.AddCommand(instancePoolCreateCmd)
}
