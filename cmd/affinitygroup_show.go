package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// affinitygroupShowCmd represents the affinitygroup show command
var affinitygroupShowCmd = &cobra.Command{
	Use:     "show <name | id>",
	Short:   "Show affinity group details",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		return showAffinityGroup(args[0])
	},
}

func showAffinityGroup(name string) error {
	ag, err := getAffinityGroupByName(name)
	if err != nil {
		return err
	}

	if len(ag.VirtualMachineIDs) == 0 {
		return nil
	}

	resp, err := cs.ListWithContext(gContext, &egoscale.ListVirtualMachines{IDs: ag.VirtualMachineIDs})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Instance Display Name", "Instance ID"})

	for _, r := range resp {
		vm := r.(*egoscale.VirtualMachine)
		table.Append([]string{vm.Name, vm.ID.String()})
	}

	table.Render()

	return nil
}

func init() {
	affinitygroupCmd.AddCommand(affinitygroupShowCmd)
}
