package cmd

import (
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var eipShowCmd = &cobra.Command{
	Use:     "show <ip address | eip id>",
	Short:   "Show an eip details",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		ip, vms, err := eipDetails(args[0])
		if err != nil {
			return err
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"Zone", "IP", "Virtual Machine", "Virtual Machine ID"})

		zone := ip.ZoneName
		ipaddr := ip.IPAddress.String()
		if len(vms) > 0 {
			for _, vm := range vms {
				table.Append([]string{zone, ipaddr, vm.Name, vm.ID})
				zone = ""
				ipaddr = ""
			}
		} else {
			table.Append([]string{zone, ipaddr})
		}
		table.Render()
		return nil
	},
}

func eipDetails(eip string) (*egoscale.IPAddress, []egoscale.VirtualMachine, error) {

	var eipID = eip
	if !isUUID(eip) {
		var err error
		eipID, err = getEIPIDByIP(cs, eip)
		if err != nil {
			return nil, nil, err
		}
	}

	addr := &egoscale.IPAddress{ID: eipID, IsElastic: true}
	if err := cs.GetWithContext(gContext, addr); err != nil {
		return nil, nil, err
	}

	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{ZoneID: addr.ZoneID})
	if err != nil {
		return nil, nil, err
	}

	vmAssociated := []egoscale.VirtualMachine{}

	for _, value := range vms {
		vm := value.(*egoscale.VirtualMachine)
		nic := vm.DefaultNic()
		if nic == nil {
			continue
		}
		for _, sIP := range nic.SecondaryIP {
			if sIP.IPAddress.Equal(addr.IPAddress) {
				vmAssociated = append(vmAssociated, *vm)
			}
		}
	}

	return addr, vmAssociated, nil
}

func init() {
	eipCmd.AddCommand(eipShowCmd)
}
