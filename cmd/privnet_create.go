package cmd

import (
	"fmt"
	"net"
	"strconv"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var privnetCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Short:   "Create private network",
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		name := args[0]

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

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zone == "" {
			zone = gCurrentAccount.DefaultZone
		}

		if isEmptyArgs(name, zone) {
			return cmd.Usage()
		}

		newNet, err := privnetCreate(name, desc, zone, startIP.Value(), endIP.Value(), netmask.Value())
		if err != nil {
			return err
		}

		return privnetShow(*newNet)
	},
}

func isEmptyArgs(args ...string) bool {
	for _, arg := range args {
		if arg == "" {
			return true
		}
	}
	return false
}

func privnetCreate(name, desc, zoneName string, startIP, endIP, netmask net.IP) (*egoscale.Network, error) {
	zone, err := getZoneByName(zoneName)

	if err != nil {
		return nil, err
	}

	if startIP != nil && endIP != nil && netmask == nil {
		netmask = net.IPv4(255, 255, 255, 0)
	}

	req := &egoscale.CreateNetwork{
		Name:        name,
		DisplayText: desc,
		ZoneID:      zone.ID,
		StartIP:     startIP,
		EndIP:       endIP,
		Netmask:     netmask,
	}

	resp, err := cs.RequestWithContext(gContext, req)
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Network), nil
}

func init() {
	privnetCreateCmd.Flags().StringP("description", "d", "", "Private network description")
	privnetCreateCmd.Flags().StringP("cidrmask", "c", "", "the cidrmask of the network. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("zone", "z", "", "Assign private network to a zone")

	startIP := new(ipValue)
	endIP := new(ipValue)
	netmask := new(ipValue)

	privnetCreateCmd.Flags().VarP(startIP, "startip", "s", "the beginning IP address in the network IP range. Required for managed networks.")
	privnetCreateCmd.Flags().VarP(endIP, "endip", "e", "the ending IP address in the network IP range. Required for managed networks.")
	privnetCreateCmd.Flags().VarP(netmask, "netmask", "m", "the netmask of the network. Required for managed networks.")

	privnetCmd.AddCommand(privnetCreateCmd)
}
