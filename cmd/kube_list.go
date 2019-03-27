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
	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{
		Tags: []egoscale.ResourceTag{{
			Key:   "managedby",
			Value: "exokube",
		}},
	})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "IP Address", "Size", "Version", "Calico", "Docker", "State"})

	for _, key := range vms {
		vm := key.(*egoscale.VirtualMachine)

		table.Append([]string{
			vm.Name,
			vm.IP().String(),
			vm.ServiceOfferingName,
			getKubeInstanceVersion(vm, kubeTagKubernetes),
			getKubeInstanceVersion(vm, kubeTagCalico),
			getKubeInstanceVersion(vm, kubeTagDocker),
			vm.State})
	}

	table.Render()

	return nil
}

func init() {
	kubeCmd.AddCommand(kubeListCmd)
}
