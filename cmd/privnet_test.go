package cmd

import (
	"net"
	"testing"
)

type dhcp struct {
	startIP net.IP
	endIP   net.IP
	netmask net.IP
}

func TestDhcpRange(t *testing.T) {
	cases := []struct {
		in   dhcp
		want string
	}{
		{
			in: dhcp{
				startIP: net.ParseIP("10.0.0.2"),
				endIP:   net.ParseIP("10.0.0.100"),
				netmask: net.ParseIP("255.255.255.0"),
			},
			want: "10.0.0.2-10.0.0.100 /24",
		},
		{
			in: dhcp{
				startIP: net.ParseIP("10.0.0.1"),
				endIP:   net.ParseIP("10.0.0.200"),
				netmask: net.ParseIP("255.255.0.0"),
			},
			want: "10.0.0.1-10.0.0.200 /16",
		},
		{
			in: dhcp{
				startIP: net.ParseIP("10.0.0.1"),
				endIP:   net.ParseIP("foo"),
				netmask: net.ParseIP("255.255.0.0"),
			},
			want: "n/a",
		},
	}
	for _, c := range cases {
		result := dhcpRange(c.in.startIP, c.in.endIP, c.in.netmask)
		if result != c.want {
			t.Errorf("got %s, want %s", result, c.want)
		}
	}
}
