package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolCreateItemOutput struct {
	ID              string   `json:"id"`
	Name            string   `json:"description"`
	Description     string   `json:"name"`
	Serviceoffering string   `json:"serviceoffering"`
	Template        string   `json:"template"`
	Zone            string   `json:"zone"`
	Affinitygroups  []string `json:"affinitygroups"`
	Securitygroups  []string `json:"securitygroups"`
	Privnets        []string `json:"Privnets"`
	Keypair         string   `json:"keypair"`
	Size            int64    `json:"size"`
	State           string   `json:"state"`
}

func (o *instancePoolCreateItemOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolCreateItemOutput) toText()  { outputText(o) }
func (o *instancePoolCreateItemOutput) toTable() { outputTable(o) }

var instancePoolCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "Create an instance pool",
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
			userData, err = getUserData(userDataPath)
			if err != nil {
				return err
			}

			if len(userData) >= maxUserDataLength {
				return fmt.Errorf("user-data maximum allowed length is %d bytes", maxUserDataLength)
			}
		}

		r := asyncTasks([]task{task{
			egoscale.CreateInstancePool{
				Name:              args[0],
				Description:       description,
				ZoneID:            zone.ID,
				ServiceOfferingID: servOffering.ID,
				TemplateID:        template.ID,
				KeyPair:           keypair,
				Size:              size,
				SecurityGroupIDs:  securityGroups,
				NetworkIDs:        privnets,
				UserData:          userData,
			},
			fmt.Sprintf("Creating instance pool %q", args[0]),
		}})
		errs := filterErrors(r)
		if len(errs) > 0 {
			return errs[0]
		}
		pool := r[0].resp.(*egoscale.CreateInstancePoolResponse)

		if !gQuiet {
			return showInstancePool(pool.ID.String())
		}

		return nil

	},
}

func init() {
	// Required Flags

	instancePoolCreateCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCreateCmd.MarkFlagRequired("zone")
	instancePoolCreateCmd.Flags().StringP("service-offering", "o", "", serviceOfferingHelp)
	instancePoolCreateCmd.MarkFlagRequired("service-offering")
	instancePoolCreateCmd.Flags().StringP("template", "t", "", "Instance pool template")
	instancePoolCreateCmd.MarkFlagRequired("template")
	instancePoolCreateCmd.Flags().IntP("size", "", 0, "Number of instance in the pool")
	instancePoolCreateCmd.MarkFlagRequired("size")

	instancePoolCreateCmd.Flags().StringP("description", "d", "", "Instance pool description")
	instancePoolCreateCmd.Flags().StringP("cloud-init", "c", "", "Cloud-init file path")
	instancePoolCreateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolCreateCmd.Flags().StringP("keypair", "k", "", "Instance pool keypair")
	instancePoolCreateCmd.Flags().StringSliceP("security-group", "s", nil, "Security groups <name | id, name | id, ...>")
	instancePoolCreateCmd.Flags().StringSliceP("privnet", "p", nil, "Privnets <name | id, name | id, ...>")
	instancePoolCmd.AddCommand(instancePoolCreateCmd)
}
