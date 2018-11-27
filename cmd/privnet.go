package cmd

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// privnetCmd represents the pn command
var privnetCmd = &cobra.Command{
	Use:   "privnet",
	Short: "Private networks management",
}

func getNetworkByName(name string) (*egoscale.Network, error) {
	net := &egoscale.Network{
		Type:            "Isolated",
		CanUseForDeploy: true,
	}

	id, errUUID := egoscale.ParseUUID(name)
	if errUUID != nil {
		net.Name = name
	} else {
		net.ID = id
	}

	resp, err := cs.GetWithContext(gContext, net)
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Network), nil
}

func getNetworkByNameAndZone(name string, zoneID *egoscale.UUID) (*egoscale.Network, error) {
	net := &egoscale.Network{
		ZoneID:          zoneID,
		Type:            "Isolated",
		CanUseForDeploy: true,
	}

	id, errUUID := egoscale.ParseUUID(name)
	if errUUID != nil {
		net.Name = name
	} else {
		net.ID = id
	}

	resp, err := cs.GetWithContext(gContext, net)
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Network), nil
}
func init() {
	RootCmd.AddCommand(privnetCmd)
}

// dhcpRange returns the string representation for a DHCP
func dhcpRange(network egoscale.Network) string {
	if network.StartIP != nil && network.EndIP != nil && network.Netmask != nil {
		mask := (net.IPMask)(network.Netmask.To4())
		ones, _ := mask.Size()
		return fmt.Sprintf("%s-%s /%d", network.StartIP, network.EndIP, ones)
	}
	return "n/a"
}

// dhcpRange returns the string representation for a DHCP
func nicIP(nic egoscale.Nic) string {
	ip := nic.IPAddress
	if ip != nil {
		return ip.String()
	}
	return "n/a"
}
