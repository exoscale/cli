package cmd

import (
	"bytes"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	privnetCmd.AddCommand(&cobra.Command{
		Use:   "show <privnet name | id>",
		Short: "Show a private network details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}
			return showPrivnet(args[0])
		},
	})
}

func showPrivnet(name string) error {
	network, err := getNetworkByName(name)
	if err != nil {
		return err
	}

	vms, err := privnetDetails(network)
	if err != nil {
		return err
	}

	dhcp := dhcpRange(*network)

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{network.Name})

	t.Append([]string{"ID", network.ID.String()})
	t.Append([]string{"Name", network.Name})
	t.Append([]string{"Description", network.DisplayText})
	t.Append([]string{"Zone", network.ZoneName})
	t.Append([]string{"DHCP IP Range", dhcp})

	if len(vms) == 0 {
		t.Append([]string{"Instances", "n/a"})
		t.Render()
		return nil
	}

	if len(vms) > 0 {
		buf := bytes.NewBuffer(nil)
		it := table.NewEmbeddedTable(buf)
		it.SetHeader([]string{" "})
		for _, vm := range vms {
			it.Append([]string{vm.Name, vm.ID.String()})
		}
		it.Render()
		t.Append([]string{"Instances", buf.String()})
	}

	t.Render()

	return nil
}

func privnetDetails(network *egoscale.Network) ([]egoscale.VirtualMachine, error) {
	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{
		ZoneID: network.ZoneID,
	})
	if err != nil {
		return nil, err
	}

	var vmsRes []egoscale.VirtualMachine
	for _, v := range vms {
		vm := v.(*egoscale.VirtualMachine)

		nic := vm.NicByNetworkID(*network.ID)
		if nic != nil {
			vmsRes = append(vmsRes, *vm)
		}
	}

	return vmsRes, nil
}
