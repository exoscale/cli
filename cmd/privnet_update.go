package cmd

import (
	"fmt"
	"net"
	"strconv"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

var privnetUpdateCmd = &cobra.Command{
	Use:   "update NAME|ID",
	Short: "Update a Private Network",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		network, err := getNetwork(args[0], nil)
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

		if netmask.IP != nil && cidrmask != "" {
			return fmt.Errorf("--netmask %q and --cidrmask %q are mutually exclusive", netmask, cidrmask)
		}

		if cidrmask != "" {
			c, err := strconv.ParseInt(cidrmask, 10, 32)
			if err != nil {
				return err
			}

			ipmask := net.CIDRMask(int(c), 32)
			netmask.IP = (*net.IP)(&ipmask)
		}

		updatedPrivnet, err := updatePrivnet(id, name, desc, startIP.Value(), endIP.Value(), netmask.Value())
		if err != nil {
			return err
		}

		return output(showPrivnet(updatedPrivnet))
	},
}

func updatePrivnet(id, name, desc string, startIP, endIP, netmask net.IP) (*egoscale.Network, error) {
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

	resp, err := asyncRequest(req, fmt.Sprintf("Updating the network %s", id))
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Network), nil
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
	privnetUpdateCmd.Flags().VarP(netmask, "netmask", "m", "the netmask of the network.")

	privnetCmd.AddCommand(privnetUpdateCmd)
}
