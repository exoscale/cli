package cmd

import (
	"fmt"
	"net"
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var privnetUpdateCmd = &cobra.Command{
	Use:   "update <name | id>",
	Short: "Update a private network",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		network, err := getNetworkByName(args[0])
		if err != nil {
			return err
		}
		id := network.ID.String()
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		startip, err := cmd.Flags().GetString("startip")
		if err != nil {
			return err
		}
		endip, err := cmd.Flags().GetString("endip")
		if err != nil {
			return err
		}
		netmask, err := cmd.Flags().GetString("netmask")
		if err != nil {
			return err
		}
		return privnetUpdate(id, name, desc, startip, endip, netmask)
	},
}

func privnetUpdate(id, name, desc, startIPAddr, endIPAddr, netmask string) error {
	uuid, err := egoscale.ParseUUID(id)
	if err != nil {
		return fmt.Errorf("update the network by ID, got %q", id)
	}

	startIP := net.ParseIP(startIPAddr)
	endIP := net.ParseIP(endIPAddr)
	netmaskIP := net.ParseIP(netmask)

	req := &egoscale.UpdateNetwork{
		ID:          uuid,
		DisplayText: desc,
		Name:        name,
		StartIP:     startIP,
		EndIP:       endIP,
		Netmask:     netmaskIP,
	}

	resp, err := asyncRequest(req, fmt.Sprintf("Updating the network %q ", id))
	if err != nil {
		return err
	}

	newNet := resp.(*egoscale.Network)

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "ID", "DHCP"})
	table.Append([]string{
		newNet.Name,
		newNet.DisplayText,
		newNet.ID.String(),
		dhcpRange(newNet.StartIP, newNet.EndIP, newNet.Netmask)})
	table.Render()
	return nil
}

func init() {
	privnetUpdateCmd.Flags().StringP("name", "n", "", "Private network name")
	privnetUpdateCmd.Flags().StringP("description", "d", "", "Private network description")
	privnetUpdateCmd.Flags().StringP("startip", "s", "", "the beginning IP address in the network IP range. Required for managed networks.")
	privnetUpdateCmd.Flags().StringP("endip", "e", "", "the ending IP address in the network IP range. Required for managed networks.")
	privnetUpdateCmd.Flags().StringP("netmask", "m", "", "the netmask of the network.  Required for managed networks")
	privnetCmd.AddCommand(privnetUpdateCmd)
}
