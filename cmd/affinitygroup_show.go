package cmd

import (
	"bytes"
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

	buf := bytes.NewBuffer(nil)
	it := table.NewEmbeddedTable(buf)
	it.SetHeader([]string{" "})
	for _, r := range resp {
		vm := r.(*egoscale.VirtualMachine)
		it.Append([]string{vm.Name, vm.ID.String()})
	}
	it.Render()
	t.Append([]string{"Instances", buf.String()})

	t.Render()

	return nil
}

func init() {
	affinitygroupCmd.AddCommand(affinitygroupShowCmd)
}
