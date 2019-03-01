package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	fmt.Fprintf(w, "Affinity Group:\t%s\n", ag.Name)     // nolint: errcheck
	fmt.Fprintf(w, "ID:\t%s\n", ag.ID)                   // nolint: errcheck
	fmt.Fprintf(w, "Description:\t%s\n", ag.Description) // nolint: errcheck
	fmt.Fprintf(w, "Type:\t%s\n", ag.Type)               // nolint: errcheck

	if len(ag.VirtualMachineIDs) == 0 {
		fmt.Fprintf(w, "VirtualMachines:\tn/a\n") // nolint: errcheck
		w.Flush()
		return nil
	}

	fmt.Fprintf(w, "VirtualMachines:\n") // nolint: errcheck
	w.Flush()

	resp, err := cs.ListWithContext(gContext, &egoscale.ListVirtualMachines{IDs: ag.VirtualMachineIDs})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "ID"})

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
