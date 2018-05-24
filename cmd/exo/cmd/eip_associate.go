package cmd

import (
	"fmt"
	"log"
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
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Usage()
			return
		}

		if err := associateIP(args[0], args[1]); err != nil {
			log.Fatal(err)
		}
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

	result := resp.(*egoscale.AddIPToNicResponse)

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Virtual machine", "IP", "Secondary IP"})

	table.Append([]string{vm.Name, vm.IP().String(), result.NicSecondaryIP.IPAddress.String()})

	table.Render()

	return nil
}

func init() {
	eipCmd.AddCommand(eipAssociateCmd)
}
