package cmd

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var privnetUpdateCmd = &cobra.Command{
	Use:   "update <name | id> [flags]",
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

		startIP, err := getIPValue(cmd, "startip")
		if err != nil {
			return err
		}

		endIP, err := getIPValue(cmd, "endip")
		if err != nil {
			return err
		}

		netmask, err := getIPValue(cmd, "netmask")
		if err != nil {
			return err
		}

		cidrmask, err := cmd.Flags().GetString("cidrmask")
		if err != nil {
			return err
		}

		if netmask.Value() != nil && cidrmask != "" {
			return fmt.Errorf("netmask %q and cidrmask %q are mutually exclusive", netmask, cidrmask)
		}

		if cidrmask != "" {
			c, err := strconv.ParseInt(cidrmask, 10, 32)
			if err != nil {
				return err
			}

			ipmask := net.CIDRMask(int(c), 32)
			netmask.IP = (*net.IP)(&ipmask)
		}

		newNet, err := privnetUpdate(id, name, desc, startIP.Value(), endIP.Value(), netmask.Value())
		if err != nil {
			return err
		}

		return privnetShow(*newNet)
	},
}

func privnetUpdate(id, name, desc string, startIP, endIP, netmask net.IP) (*egoscale.Network, error) {
	uuid, err := egoscale.ParseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("update the network by ID, got %q", id)
	}

	if startIP != nil && endIP != nil && netmask == nil {
		netmask = net.IPv4(255, 255, 255, 0)
	}

	req := &egoscale.UpdateNetwork{
		ID:          uuid,
		DisplayText: desc,
		Name:        name,
		StartIP:     startIP,
		EndIP:       endIP,
		Netmask:     netmask,
	}

	resp, err := asyncRequest(req, fmt.Sprintf("Updating the network %q ", id))
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Network), nil
}

func privnetShow(network egoscale.Network) error {
	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "ID", "DHCP"})
	table.Append([]string{
		network.Name,
		network.DisplayText,
		network.ID.String(),
		dhcpRange(network)})
	table.Render()
	return nil
}

func init() {
	privnetUpdateCmd.Flags().StringP("name", "n", "", "Private network name")
	privnetUpdateCmd.Flags().StringP("description", "d", "", "Private network description")
	privnetUpdateCmd.Flags().StringP("cidrmask", "c", "", "the cidrmask of the network. E.g 32")

	startIP := new(ipValue)
	endIP := new(ipValue)
	netmask := new(ipValue)

	privnetUpdateCmd.Flags().VarP(startIP, "startip", "s", "the beginning IP address in the network IP range.")
	privnetUpdateCmd.Flags().VarP(endIP, "endip", "e", "the ending IP address in the network IP range.")
	privnetUpdateCmd.Flags().VarP(netmask, "netmask", "n", "the netmask of the network.")

	privnetCmd.AddCommand(privnetUpdateCmd)
}
