package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
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

		id, err := egoscale.ParseUUID(args[0])
		if err != nil {
			id, err = getEIPIDByIP(args[0])
			if err != nil {
				return err
			}
		}

		ip, vms, err := eipDetails(id)
		if err != nil {
			return err
		}

		table := table.NewTable(os.Stdout)

		zone := ip.ZoneName
		ipaddr := ip.IPAddress.String()

		table.SetHeader([]string{ipaddr})
		table.Append([]string{"ID", id.String()})
		table.Append([]string{"Zone", zone})
		if ip.Healthcheck != nil {
			table.Append([]string{
				"Healthcheck Mode",
				ip.Healthcheck.Mode})
			table.Append([]string{
				"Healthcheck Path",
				ip.Healthcheck.Path})
			table.Append([]string{
				"Healthcheck Port",
				fmt.Sprintf("%d", ip.Healthcheck.Port)})
			table.Append([]string{
				"Healthcheck Interval",
				fmt.Sprintf("%d", ip.Healthcheck.Interval)})
			table.Append([]string{
				"Healthcheck Strikes Fail",
				fmt.Sprintf("%d", ip.Healthcheck.StrikesFail)})
			table.Append([]string{
				"Healthcheck Strikes OK",
				fmt.Sprintf("%d", ip.Healthcheck.StrikesOk)})
			table.Append([]string{
				"Healthcheck Timeout",
				fmt.Sprintf("%d", ip.Healthcheck.Timeout)})
		}

		if len(vms) > 0 {
			table.Append([]string{
				"Instances",
				vms[0].Name,
			})
			for _, vm := range vms[1:] {
				table.Append([]string{
					"",
					vm.Name})
			}
		}
		table.Render()
		return nil
	},
}

func eipDetails(eip *egoscale.UUID) (*egoscale.IPAddress, []egoscale.VirtualMachine, error) {
	var eipID = eip

	query := &egoscale.IPAddress{ID: eipID, IsElastic: true}
	resp, err := cs.GetWithContext(gContext, query)
	if err != nil {
		return nil, nil, err
	}

	addr := resp.(*egoscale.IPAddress)
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
