package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolCreateItemOutput egoscale.CreateInstancePoolResponse

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

		if zoneName == "" {
			zoneName = gCurrentAccount.DefaultZone
		}

		zone, err := getZoneIDByName(zoneName)
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

		if templateName == "" {
			templateName = gCurrentAccount.DefaultTemplate
		}

		template, err := getTemplateByName(zone, templateName, templateFilter)
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

		//It use asyncTasks to have spinner when user exec this command
		r := asyncTasks([]task{task{
			egoscale.CreateInstancePool{
				Name:              args[0],
				Description:       description,
				ZoneID:            zone,
				ServiceofferingID: servOffering.ID,
				TemplateID:        template.ID,
				Keypair:           keypair,
				Size:              size,
				// SecuritygroupIDs:  []egoscale.UUID{*egoscale.MustParseUUID("a4430e9f-11e3-4da4-bf78-325d3640ea17")},
				// AffinitygroupIDs:  []egoscale.UUID{*egoscale.MustParseUUID("9cf2a2fb-31e2-4ead-9c22-2339e1c71fbc")},
				// NetworkIDs:        []egoscale.UUID{*egoscale.MustParseUUID("a7089285-98f2-98c3-c3a9-786b8a4c1ab4")},
			},
			fmt.Sprintf("Create instance pool %q", args[0]),
		}})
		errs := filterErrors(r)
		if len(errs) > 0 {
			return errs[0]
		}
		pool := r[0].resp.(*egoscale.CreateInstancePoolResponse)
		o := instancePoolCreateItemOutput(*pool)

		return output(&o, nil)
	},
}

func init() {
	instancePoolCreateCmd.Flags().StringP("description", "d", "", "Instance pool description")
	instancePoolCreateCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCreateCmd.Flags().StringP("service-offering", "s", "small", "Instance pool service offering")
	instancePoolCreateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolCreateCmd.Flags().StringP("template", "t", "", "Instance pool template")
	instancePoolCreateCmd.Flags().StringP("keypair", "k", "", "Instance pool keypair")
	instancePoolCreateCmd.Flags().IntP("size", "", 2, "Number of instance in the pool")
	instancePoolCmd.AddCommand(instancePoolCreateCmd)
}
