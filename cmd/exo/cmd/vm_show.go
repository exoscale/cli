package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var vmShowCmd = &cobra.Command{
	Use:   "show <name | id>",
	Short: "show detailed information of a virtual machine",
}

func vmShowCmdRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		vmShowCmd.Usage()
		return
	}
	if err := showVM(args[0]); err != nil {
		log.Fatal(err)
	}
}

func showVM(name string) error {
	vm, err := getVMWithNameOrID(cs, name)
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{vm.Name})

	table.Append([]string{"OS Template", vm.TemplateName})

	table.Append([]string{"Region", vm.ZoneName})

	temp := &egoscale.Template{IsFeatured: true, ID: vm.TemplateID, ZoneID: "1"}

	if err := cs.Get(temp); err != nil {
		return err
	}
	table.Append([]string{"Instance Type", vm.ServiceOfferingName})

	table.Append([]string{"Disk", fmt.Sprintf("%d GB", temp.Size>>30)})

	table.Append([]string{"Instance Hostname", vm.Name})

	table.Append([]string{"Instance Display Name", vm.DisplayName})

	table.Append([]string{"Created on", vm.Created})

	table.Append([]string{"Base SSH Key", vm.KeyPair})

	sgs := getSecurityGroup(vm)

	sgName := strings.Join(sgs, " - ")

	table.Append([]string{"Security Group", sgName})

	table.Append([]string{"Instance IP", vm.IP().String()})

	table.Render()

	return nil
}

func init() {
	vmShowCmd.Run = vmShowCmdRun
	vmCmd.AddCommand(vmShowCmd)
}
