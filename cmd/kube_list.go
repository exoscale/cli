package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// kubeListCmd represents the kube list command
var kubeListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List running Kubernetes cluster instances",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listKubeInstances()
	},
}

func listKubeInstances() error {
	/* FIXME: with egoscale >= 0.13, replace this implementation with:
	   vms, err := cs.ListWithContext(gContext, &egoscale.ListVirtualMachines{
	       Tag: []egoscale.ResourceTag{{
	           Key: "managedby",
	           Value: "exokube",
	       }}
	   })
	*/
	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "IP Address", "Size", "Version", "State"})

	for _, key := range vms {
		vm := key.(*egoscale.VirtualMachine)

		kubeVersion := getKubeInstanceVersion(vm)
		if kubeVersion == "" {
			continue
		}

		table.Append([]string{vm.Name, vm.IP().String(), vm.ServiceOfferingName, kubeVersion, vm.State})
	}

	table.Render()

	return nil
}

func init() {
	kubeCmd.AddCommand(kubeListCmd)
}
