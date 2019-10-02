package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolItemOutput struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	Serviceoffering string                     `json:"serviceoffering"`
	Template        string                     `json:"templateid"`
	Zone            string                     `json:"zoneid"`
	Affinitygroups  []string                   `json:"affinitygroups"`
	Securitygroups  []string                   `json:"securitygroups"`
	Privnets        []string                   `json:"Privnets"`
	Keypair         string                     `json:"keypair"`
	Size            int                        `json:"size"`
	State           egoscale.InstancePoolState `json:"state"`
	Virtualmachines []string                   `json:"virtualmachines"`
}

func (o *instancePoolItemOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolItemOutput) toText()  { outputText(o) }
func (o *instancePoolItemOutput) toTable() { outputTable(o) }

var instancePoolShowCmd = &cobra.Command{
	Use:     "show <name | id>",
	Short:   "Create an instance pool",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
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

		instancePool, err := getInstancePoolByName(args[0], zone.ID)
		if err != nil {
			return err
		}

		so, err := getServiceOfferingByName(instancePool.ServiceofferingID.String())
		if err != nil {
			return err
		}

		template, err := getTemplateByName(instancePool.ZoneID, instancePool.TemplateID.String(), "featured")
		if err != nil {
			template, err = getTemplateByName(instancePool.ZoneID, instancePool.TemplateID.String(), "self")
			if err != nil {
				return err
			}
		}

		o := instancePoolItemOutput{
			ID:              instancePool.ID.String(),
			Name:            instancePool.Name,
			Description:     instancePool.Description,
			Serviceoffering: so.Name,
			Template:        template.Name,
			Zone:            zone.Name,
			Keypair:         instancePool.Keypair,
			Size:            instancePool.Size,
			State:           instancePool.State,
		}
		for _, vm := range instancePool.Virtualmachines {
			o.Virtualmachines = append(o.Virtualmachines, vm.Name)
		}
		for _, a := range instancePool.AffinitygroupIDs {
			aff, err := getAffinityGroupByName(a.String())
			if err != nil {
				return err
			}
			o.Affinitygroups = append(o.Affinitygroups, aff.Name)
		}
		for _, s := range instancePool.SecuritygroupIDs {
			sg, err := getSecurityGroupByNameOrID(s.String())
			if err != nil {
				return err
			}
			o.Securitygroups = append(o.Securitygroups, sg.Name)
		}
		for _, i := range instancePool.NetworkIDs {
			net, err := getNetwork(i.String(), instancePool.ZoneID)
			if err != nil {
				return err
			}
			name := net.Name
			if name == "" {
				name = net.ID.String()
			}
			o.Privnets = append(o.Privnets, name)
		}

		return output(&o, err)
	},
}

func init() {
	instancePoolShowCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCmd.AddCommand(instancePoolShowCmd)
}
