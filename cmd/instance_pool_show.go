package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolItemOutput struct {
	ID                *egoscale.UUID             `json:"id"`
	Name              string                     `json:"name"`
	Description       string                     `json:"description"`
	ServiceofferingID *egoscale.UUID             `json:"serviceofferingid"`
	TemplateID        *egoscale.UUID             `json:"templateid"`
	ZoneID            *egoscale.UUID             `json:"zoneid"`
	AffinitygroupIDs  []egoscale.UUID            `json:"affinitygroupids"`
	SecuritygroupIDs  []egoscale.UUID            `json:"securitygroupids"`
	NetworkIDs        []egoscale.UUID            `json:"networkids"`
	Keypair           string                     `json:"keypair"`
	Size              int                        `json:"size"`
	State             egoscale.InstancePoolState `json:"state"`
	Virtualmachines   []string                   `json:"virtualmachines"`
}

func (o *instancePoolItemOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolItemOutput) toText()  { outputText(o) }
func (o *instancePoolItemOutput) toTable() { outputTable(o) }

var instancePoolShowCmd = &cobra.Command{
	Use:     "delete <name | id>",
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

		zone, err := getZoneIDByName(zoneName)
		if err != nil {
			return err
		}

		i, err := getInstancePoolByName(args[0], zone)
		if err != nil {
			return err
		}

		o := instancePoolItemOutput{
			ID:                i.ID,
			Name:              i.Name,
			Description:       i.Description,
			ServiceofferingID: i.ServiceofferingID,
			TemplateID:        i.TemplateID,
			ZoneID:            i.ZoneID,
			AffinitygroupIDs:  i.AffinitygroupIDs,
			SecuritygroupIDs:  i.SecuritygroupIDs,
			NetworkIDs:        i.NetworkIDs,
			Keypair:           i.Keypair,
			Size:              i.Size,
			State:             i.State,
		}
		for _, vm := range i.Virtualmachines {
			o.Virtualmachines = append(o.Virtualmachines, vm.ID.String())
		}

		return output(&o, err)
	},
}

func init() {
	instancePoolShowCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCmd.AddCommand(instancePoolShowCmd)
}
