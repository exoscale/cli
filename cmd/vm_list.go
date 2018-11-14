package cmd

import (
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var vmlistCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all the virtual machines instances",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return cmd.Usage()
		}
		return listVMs()
	},
}

func listVMs() error {
	vm := &egoscale.VirtualMachine{}
	vms, err := cs.ListWithContext(gContext, vm)
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Security Group", "IP Address", "Status", "Zone", "ID"})

	for _, key := range vms {
		vm := key.(*egoscale.VirtualMachine)

		sgs := getSecurityGroup(vm)

		sgName := strings.Join(sgs, " - ")
		table.Append([]string{vm.Name, sgName, vm.IP().String(), vm.State, vm.ZoneName, vm.ID.String()})
	}
	table.Render()

	return nil
}

func init() {
	vmCmd.AddCommand(vmlistCmd)
}
