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

		sip, err := cmd.Flags().GetString("startip")
		if err != nil {
			return err
		}
		startip := net.ParseIP(sip)

		eip, err := cmd.Flags().GetString("endip")
		if err != nil {
			return err
		}
		endip := net.ParseIP(eip)

		nmask, err := cmd.Flags().GetString("netmask")
		if err != nil {
			return err
		}

		cidrmask, err := cmd.Flags().GetString("cidrmask")
		if err != nil {
			return err
		}

		if nmask != "" && cidrmask != "" {
			return fmt.Errorf("netmask %q and cidrmask %q are mutually exclusive", nmask, cidrmask)
		}

		netmask := net.ParseIP(nmask)
		if netmask == nil {
			if cidrmask != "" {
				c, err := strconv.ParseInt(cidrmask, 10, 32)
				if err != nil {
					return err
				}

				ipmask := net.CIDRMask(int(c), 32)
				netmask = (net.IP)(ipmask)
			}
		}

		newNet, err := privnetUpdate(id, name, desc, startip, endip, netmask)
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
	privnetUpdateCmd.Flags().StringP("startip", "s", "", "the beginning IP address in the network IP range. Required for managed networks.")
	privnetUpdateCmd.Flags().StringP("endip", "e", "", "the ending IP address in the network IP range. Required for managed networks.")
	privnetUpdateCmd.Flags().StringP("netmask", "m", "", "the netmask of the network. E.g. 255.255.255.0")
	privnetUpdateCmd.Flags().StringP("cidrmask", "c", "", "the cidrmask of the network. E.g 32")
	privnetCmd.AddCommand(privnetUpdateCmd)
}
