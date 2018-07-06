package cmd

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// eipCmd represents the eip command
var eipCmd = &cobra.Command{
	Use:   "eip",
	Short: "Elastic IPs management",
}

func getEIPIDByIP(cs *egoscale.Client, ipAddr string) (string, error) {
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address %q", ipAddr)
	}
	eips, err := cs.List(&egoscale.IPAddress{IsElastic: true})
	if err != nil {
		return "", err
	}

	for _, e := range eips {
		eip := e.(*egoscale.IPAddress)
		if eip.IPAddress.Equal(ip) {
			return eip.ID, nil
		}
	}

	return "", fmt.Errorf("elastic IP %q not found", ipAddr)
}

func init() {
	RootCmd.AddCommand(eipCmd)
}
