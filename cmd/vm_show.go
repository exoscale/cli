package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var vmShowCmd = &cobra.Command{
	Use:     "show <name | id>",
	Short:   "Show virtual machine details",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		return showVM(args[0])
	},
}

func showVM(name string) error {
	vm, err := getVirtualMachineByNameOrID(name)
	if err != nil {
		return err
	}

	resp, err := cs.GetWithContext(gContext, &egoscale.Template{
		IsFeatured: true,
		ID:         vm.TemplateID,
		ZoneID:     vm.ZoneID,
	})

	if err != nil {
		return err
	}

	temp := resp.(*egoscale.Template)

	resp, err = cs.GetWithContext(gContext, &egoscale.Volume{
		VirtualMachineID: vm.ID,
		Type:             "ROOT",
	})

	if err != nil {
		return err
	}

	volume := resp.(*egoscale.Volume)

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{vm.Name})

	table.Append([]string{"State", vm.State})

	table.Append([]string{"OS Template", vm.TemplateName})

	table.Append([]string{"Region", vm.ZoneName})

	table.Append([]string{"Instance Type", vm.ServiceOfferingName})

	table.Append([]string{"Disk", fmt.Sprintf("%d GB", volume.Size>>30)})

	table.Append([]string{"Instance Hostname", vm.Name})

	table.Append([]string{"Instance Display Name", vm.DisplayName})

	username, ok := temp.Details["username"]
	if !ok {
		return fmt.Errorf("template %q: failed to get username", temp.Name)
	}

	table.Append([]string{"Instance Username", username})

	table.Append([]string{"Created on", vm.Created})

	table.Append([]string{"Base SSH Key", vm.KeyPair})

	sgs := getSecurityGroup(vm)

	sgName := strings.Join(sgs, " - ")

	table.Append([]string{"Security Groups", sgName})

	ags := getAffinityGroups(vm)

	table.Append([]string{"Affinity Groups", strings.Join(ags, " - ")})

	table.Append([]string{"Instance IP", vm.IP().String()})

	table.Append([]string{"ID", vm.ID.String()})

	table.Render()

	return nil
}

func init() {
	vmCmd.AddCommand(vmShowCmd)
}
