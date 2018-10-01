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

	if err := cs.GetWithContext(gContext, net); err != nil {
		return nil, err
	}

	return net, nil
}

func init() {
	RootCmd.AddCommand(privnetCmd)
}

// dhcpRange returns the string representation for a DHCP
func dhcpRange(startIP, endIP, netmask net.IP) string {
	if startIP != nil && endIP != nil && netmask != nil {
		mask := net.IPMask(netmask.To4())
		prefixSize, _ := mask.Size()
		return fmt.Sprintf("%s-%s /%d", startIP.String(), endIP.String(), prefixSize)
	}
	return "n/a"
}
