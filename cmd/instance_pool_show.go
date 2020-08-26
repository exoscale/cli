package cmd

import (
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolItemOutput struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	ServiceOffering string                     `json:"service_offering"`
	Template        string                     `json:"templateid"`
	Zone            string                     `json:"zoneid"`
	SecurityGroups  []string                   `json:"security_groups"`
	PrivateNetworks []string                   `json:"private_networks"`
	IPv6            bool                       `json:"ipv6" outputLabel:"IPv6"`
	SSHKey          string                     `json:"ssh_key"`
	Size            int                        `json:"size"`
	DiskSize        string                     `json:"disk_size"`
	State           egoscale.InstancePoolState `json:"state"`
	Instances       []string                   `json:"instances"`
}

func (o *instancePoolItemOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolItemOutput) toText()  { outputText(o) }
func (o *instancePoolItemOutput) toTable() { outputTable(o) }

var instancePoolShowCmd = &cobra.Command{
	Use:   "show <name | id>",
	Short: "Show an instance pool details",
	Long: fmt.Sprintf(`This command shows an instance pool details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolItemOutput{}), ", ")),
	Aliases: gShowAlias,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		return showInstancePool(args[0], zone)
	},
}

func showInstancePool(name, zoneName string) error {
	zone, err := getZoneByName(zoneName)
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

	template, err := getTemplateByName(instancePool.ZoneID, instancePool.TemplateID.String(), "")
	if err != nil {
		return err
	}

	o := instancePoolItemOutput{
		ID:              instancePool.ID.String(),
		Name:            instancePool.Name,
		Description:     instancePool.Description,
		ServiceOffering: serviceOffering.Name,
		Template:        template.Name,
		Zone:            zone.Name,
		SSHKey:          instancePool.KeyPair,
		Size:            instancePool.Size,
		DiskSize:        humanize.IBytes(uint64(instancePool.RootDiskSize << 30)),
		IPv6:            instancePool.IPv6,
		State:           instancePool.State,
	}
	for _, vm := range instancePool.VirtualMachines {
		o.Instances = append(o.Instances, vm.Name)
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
		o.PrivateNetworks = append(o.PrivateNetworks, name)
	}

	return output(&o, err)
}

func init() {
	instancePoolShowCmd.Flags().StringP("zone", "z", "", "Instance pool zone")
	instancePoolCmd.AddCommand(instancePoolShowCmd)
}
