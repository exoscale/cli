package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolItemOutput struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	ServiceOffering string                     `json:"serviceoffering"`
	Template        string                     `json:"templateid"`
	Zone            string                     `json:"zoneid"`
	SecurityGroups  []string                   `json:"securitygroups"`
	Privnets        []string                   `json:"Privnets"`
	KeyPair         string                     `json:"keypair"`
	Size            int                        `json:"size"`
	State           egoscale.InstancePoolState `json:"state"`
	VirtualMachines []string                   `json:"virtualmachines"`
}

func (o *instancePoolItemOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolItemOutput) toText()  { outputText(o) }
func (o *instancePoolItemOutput) toTable() { outputTable(o) }

var instancePoolShowCmd = &cobra.Command{
	Use:     "show <name | id>",
	Short:   "Show an instance pool",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		return showInstancePool(args[0])
	},
}

func showInstancePool(name string) error {
	zone, err := getZoneByName(gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	instancePool, err := getInstancePoolByName(name, zone.ID)
	if err != nil {
		return err
	}

	zone, err = getZoneByName(instancePool.ZoneID.String())
	if err != nil {
		return err
	}

	serviceOffering, err := getServiceOfferingByName(instancePool.ServiceOfferingID.String())
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
		ServiceOffering: serviceOffering.Name,
		Template:        template.Name,
		Zone:            zone.Name,
		KeyPair:         instancePool.KeyPair,
		Size:            instancePool.Size,
		State:           instancePool.State,
	}
	for _, vm := range instancePool.VirtualMachines {
		o.VirtualMachines = append(o.VirtualMachines, vm.Name)
	}

	for _, s := range instancePool.SecurityGroupIDs {
		sg, err := getSecurityGroupByNameOrID(s.String())
		if err != nil {
			return err
		}
		o.SecurityGroups = append(o.SecurityGroups, sg.Name)
	}
	if len(instancePool.SecurityGroupIDs) == 0 {
		o.SecurityGroups = append(o.SecurityGroups, "default")
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
}

func init() {
	instancePoolCmd.AddCommand(instancePoolShowCmd)
}
