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
		if netmask == nil && cidrmask != "" {
			c, err := strconv.ParseInt(cidrmask, 10, 32)
			if err != nil {
				return err
			}
			ipmask := net.CIDRMask(int(c), 32)
			netmask = (net.IP)(ipmask)
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

		newNet, err := privnetCreate(name, desc, zone, startip, endip, netmask)
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
	zoneID, err := getZoneIDByName(zoneName)
	if err != nil {
		return nil, err
	}

	// NetworkOffering are cross zones
	listReq := &egoscale.ListNetworkOfferings{
		Name:     "PrivNet",
		ZoneID:   zoneID,
		Page:     1,
		PageSize: 1,
	}

	resp, err := cs.RequestWithContext(gContext, listReq)
	if err != nil {
		return nil, err
	}

	nos := resp.(*egoscale.ListNetworkOfferingsResponse)
	if len(nos.NetworkOffering) != 1 {
		return nil, fmt.Errorf("missing Network Offering %q in %q", listReq.Name, zoneName)
	}

	if startIP != nil && endIP != nil && netmask == nil {
		netmask = net.IPv4(255, 255, 255, 0)
	}

	req := &egoscale.CreateNetwork{
		Name:              name,
		DisplayText:       desc,
		ZoneID:            zoneID,
		StartIP:           startIP,
		EndIP:             endIP,
		Netmask:           netmask,
		NetworkOfferingID: nos.NetworkOffering[0].ID,
	}

	resp, err = cs.RequestWithContext(gContext, req)
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Network), nil
}

func init() {
	privnetCreateCmd.Flags().StringP("description", "d", "", "Private network description")
	privnetCreateCmd.Flags().StringP("startip", "s", "", "the beginning IP address in the network IP range. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("endip", "e", "", "the ending IP address in the network IP range. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("netmask", "n", "", "the netmask of the network. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("cidrmask", "c", "", "the cidrmask of the network. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("zone", "z", "", "Assign private network to a zone")
	privnetCmd.AddCommand(privnetCreateCmd)
}
