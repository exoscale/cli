package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"

	"github.com/spf13/cobra"
)

// associateCmd represents the associate command
var eipAssociateCmd = &cobra.Command{
	Use:   "associate <IP address> <instance name | instance id>",
	Short: "Associate an IP to an instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		return associateIP(args[0], args[1])
	},
}

func associateIP(ipAddr, instance string) error {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return fmt.Errorf("Invalide IP address")
	}

	vm, err := getVMWithNameOrID(cs, instance)
	if err != nil {
		return err
	}

	defaultNic := vm.DefaultNic()

	if defaultNic == nil {
		return fmt.Errorf("No default NIC on this instance")
	}

	resp, err := cs.Request(&egoscale.AddIPToNic{NicID: defaultNic.ID, IPAddress: ip})
	if err != nil {
		return err
	}

	result := resp.(*egoscale.NicSecondaryIP)

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Virtual machine", "IP", "Secondary IP"})

	table.Append([]string{vm.Name, vm.IP().String(), result.IPAddress.String()})

	table.Render()

	return nil
}

func init() {
	eipCmd.AddCommand(eipAssociateCmd)
}
