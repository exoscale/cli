package cmd

import (
	"bufio"
	"net"
	"os"

	"github.com/exoscale/cli/table"
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
		cidrmask, err := cmd.Flags().GetInt("cidrmask")
		if err != nil {
			return err
		}
		if netmask == "" && cidrmask != 0 {
			ipmask := net.CIDRMask(cidrmask, 32)
			netmask = (net.IP)(ipmask).String()
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		if zone == "" {
			zone = gCurrentAccount.DefaultZone
		}

		if zone == "" {
			reader := bufio.NewReader(os.Stdin)
			if desc == "" {
				desc, err = readInput(reader, "Description", "")
				if err != nil {
					return err
				}
			}
			if zone == "" {
				zone, err = readInput(reader, "Zone", gCurrentAccount.DefaultZone)
				if err != nil {
					return err
				}
			}
		}

		if isEmptyArgs(name, zone) {
			return cmd.Usage()
		}

		return privnetCreate(name, desc, zone, startip, endip, netmask)
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

func privnetCreate(name, desc, zoneName, startIPAddr, endIPAddr, netmask string) error {
	zone, err := getZoneIDByName(zoneName)
	if err != nil {
		return err
	}

	resp, err := cs.RequestWithContext(gContext, &egoscale.ListNetworkOfferings{ZoneID: zone, Name: "PrivNet"})
	if err != nil {
		return err
	}

	startIP := net.ParseIP(startIPAddr)
	endIP := net.ParseIP(endIPAddr)
	netmaskIP := net.ParseIP(netmask)

	s := resp.(*egoscale.ListNetworkOfferingsResponse)

	req := &egoscale.CreateNetwork{
		DisplayText: desc,
		Name:        name,
		ZoneID:      zone,
		StartIP:     startIP,
		EndIP:       endIP,
		Netmask:     netmaskIP,
	}
	if len(s.NetworkOffering) > 0 {
		req.NetworkOfferingID = s.NetworkOffering[0].ID
	}

	resp, err = cs.RequestWithContext(gContext, req)
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
		dhcpRange(startIP, endIP, netmaskIP),
	})
	table.Render()
	return nil
}

func init() {
	privnetCreateCmd.Flags().StringP("description", "d", "", "Private network description")
	privnetCreateCmd.Flags().StringP("startip", "s", "", "the beginning IP address in the network IP range. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("endip", "e", "", "the ending IP address in the network IP range. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("netmask", "n", "", "the netmask of the network. Required for managed networks.")
	privnetCreateCmd.Flags().IntP("cidrmask", "c", 0, "the cidrmask of the network. Required for managed networks.")
	privnetCreateCmd.Flags().StringP("zone", "z", "", "Assign private network to a zone")
	privnetCmd.AddCommand(privnetCreateCmd)
}
