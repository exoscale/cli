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
	Userdata        string   `json:"userdata"`
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

		if templateName == "" {
			templateName = gCurrentAccount.DefaultTemplate
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

		sgs, err := getSecurityGroups(sg)
		if err != nil {
			return err
		}

		aff, err := cmd.Flags().GetStringSlice("anti-affinity-group")
		if err != nil {
			return err
		}

		affs, err := getAffinityGroup(aff)
		if err != nil {
			return err
		}

		priv, err := cmd.Flags().GetStringSlice("privnet")
		if err != nil {
			return err
		}

		privs, err := getPrivnetList(priv, zone.ID)
		if err != nil {
			return err
		}

		//It use asyncTasks to have spinner when user exec this command
		r := asyncTasks([]task{task{
			egoscale.CreateInstancePool{
				Name:              args[0],
				Description:       description,
				ZoneID:            zone.ID,
				ServiceofferingID: servOffering.ID,
				TemplateID:        template.ID,
				Keypair:           keypair,
				Size:              size,
				SecuritygroupIDs:  sgs,
				AffinitygroupIDs:  affs,
				NetworkIDs:        privs,
			},
			fmt.Sprintf("Create instance pool %q", args[0]),
		}})
		errs := filterErrors(r)
		if len(errs) > 0 {
			return errs[0]
		}
		pool := r[0].resp.(*egoscale.CreateInstancePoolResponse)
		o, err := formatInstancePoolCreateItemOutput(pool, zone)
		return output(o, err)
	},
}

func formatInstancePoolCreateItemOutput(instancePool *egoscale.CreateInstancePoolResponse, zone *egoscale.Zone) (*instancePoolCreateItemOutput, error) {
	so, err := getServiceOfferingByName(instancePool.ServiceofferingID.String())
	if err != nil {
		return nil, err
	}

	template, err := getTemplateByName(zone.ID, instancePool.TemplateID.String(), "featured")
	if err != nil {
		template, err = getTemplateByName(instancePool.ZoneID, instancePool.TemplateID.String(), "self")
		if err != nil {
			return nil, err
		}
	}

	output := &instancePoolCreateItemOutput{
		ID:              instancePool.ID.String(),
		Name:            instancePool.Name,
		Description:     instancePool.Description,
		Keypair:         instancePool.Keypair,
		Size:            instancePool.Size,
		Template:        template.Name,
		Serviceoffering: so.Name,
		Zone:            zone.Name,
	}

	for _, a := range instancePool.AffinitygroupIDs {
		aff, err := getAffinityGroupByName(a.String())
		if err != nil {
			return nil, err
		}
		output.Affinitygroups = append(output.Affinitygroups, aff.Name)
	}
	for _, s := range instancePool.SecuritygroupIDs {
		sg, err := getSecurityGroupByNameOrID(s.String())
		if err != nil {
			return nil, err
		}
		output.Securitygroups = append(output.Securitygroups, sg.Name)
	}
	for _, i := range instancePool.NetworkIDs {
		net, err := getNetwork(i.String(), instancePool.ZoneID)
		if err != nil {
			return nil, err
		}
		name := net.Name
		if name == "" {
			name = net.ID.String()
		}
		output.Privnets = append(output.Privnets, name)
	}

	return output, nil
}

func init() {
	instancePoolCreateCmd.Flags().StringP("description", "d", "", "Instance pool description")
	instancePoolCreateCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCreateCmd.Flags().StringP("service-offering", "o", "small", "Instance pool service offering")
	instancePoolCreateCmd.Flags().StringP("template-filter", "", "featured", templateFilterHelp)
	instancePoolCreateCmd.Flags().StringP("template", "t", "", "Instance pool template")
	instancePoolCreateCmd.Flags().StringP("keypair", "k", "", "Instance pool keypair")
	instancePoolCreateCmd.Flags().IntP("size", "", 2, "Number of instance in the pool")
	instancePoolCreateCmd.Flags().StringSliceP("security-group", "s", nil, "Security groups <name | id, name | id, ...>")
	instancePoolCreateCmd.Flags().StringSliceP("anti-affinity-group", "a", nil, "Anti-Affinitygroup groups <name | id, name | id, ...>")
	instancePoolCreateCmd.Flags().StringSliceP("privnet", "p", nil, "Privnets <name | id, name | id, ...>")
	instancePoolCmd.AddCommand(instancePoolCreateCmd)
}
