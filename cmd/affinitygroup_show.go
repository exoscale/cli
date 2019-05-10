package cmd

import (
	"os"
	"strings"

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

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{ag.Name})

	t.Append([]string{"ID", ag.ID.String()})
	t.Append([]string{"Name", ag.Name})
	t.Append([]string{"Description", ag.Description})

	if len(ag.VirtualMachineIDs) == 0 {
		t.Append([]string{"Instances", "n/a"})
		t.Render()
		return nil
	}

	resp, err := cs.ListWithContext(gContext, &egoscale.ListVirtualMachines{IDs: ag.VirtualMachineIDs})
	if err != nil {
		return err
	}

	instances := make([]string, len(ag.VirtualMachineIDs))
	for i, r := range resp {
		vm := r.(*egoscale.VirtualMachine)
		instances[i] = vm.ID.String() + " │ " + vm.Name
	}
	t.Append([]string{"Instances", strings.Join(instances, "\n")})

	t.Render()

	return nil
}

func init() {
	affinitygroupCmd.AddCommand(affinitygroupShowCmd)
}
