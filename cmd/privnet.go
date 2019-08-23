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

// getNetwork returns a Private Network by name or ID, and optionally a zone to restrict search to.
// Since the Exoscale API doesn't support searching by unique name, we have to list all networks and
// match results ourselves. In case the caller provides a network name and there are multiple matching
// results an error is returned.
func getNetwork(net string, zoneID *egoscale.UUID) (*egoscale.Network, error) {
	var found *egoscale.Network

	req := &egoscale.Network{
		ZoneID:          zoneID,
		Type:            "Isolated",
		CanUseForDeploy: true,
	}

	id, errUUID := egoscale.ParseUUID(net)
	if errUUID != nil {
		req.Name = net
	} else {
		req.ID = id
	}

	resp, err := cs.ListWithContext(gContext, req)
	if err != nil {
		return nil, err
	}

	for _, item := range resp {
		network := item.(*egoscale.Network)

		// If search criteria is unique ID, return first (i.e. only) match
		if id != nil && network.ID.Equal(*id) {
			return network, nil
		}

		// If search criteria is name, check that there isn't multiple networks named
		// identically before returning a match
		if network.Name == net {
			// We already found a match before -> multiple results
			if found != nil {
				return nil, fmt.Errorf("found multiple networks named %q, please specify a unique ID instead", net)
			}
			found = network
		}
	}

	if found != nil {
		return found, nil
	}

	return nil, fmt.Errorf("network %q not found", net)
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
