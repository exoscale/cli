package cmd

import (
	"net"
	"testing"

	"github.com/exoscale/egoscale"
)

func TestDhcpRange(t *testing.T) {
	cases := []struct {
		network  egoscale.Network
		expected string
	}{
		{
			egoscale.Network{
				StartIP: net.IPv4(10, 0, 0, 2),
				EndIP:   net.IPv4(10, 0, 0, 100),
				Netmask: net.IPv4(255, 255, 255, 0),
			},
			"10.0.0.2-10.0.0.100 /24",
		},
		{
			egoscale.Network{
				StartIP: net.IPv4(10, 0, 0, 1),
				EndIP:   net.IPv4(10, 0, 0, 200),
				Netmask: net.IPv4(255, 255, 0, 0),
			},
			"10.0.0.1-10.0.0.200 /16",
		},
		{
			egoscale.Network{},
			"n/a",
		},
	}
	for _, tt := range cases {
		result := dhcpRange(tt.network)
		if result != tt.expected {
			t.Errorf("got %s, want %s", result, tt.expected)
		}
	}
}
