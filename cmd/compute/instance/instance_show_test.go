package instance

import (
	"net"
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
)

func TestFormatInstanceVPCSubnet(t *testing.T) {
	tests := []struct {
		name   string
		subnet v3.InstanceVpcSubnets
		want   string
	}{
		{"ipv4 only", v3.InstanceVpcSubnets{Name: "front", Ipv4: net.ParseIP("10.0.0.5")}, "front 10.0.0.5"},
		{"ipv6 only", v3.InstanceVpcSubnets{Name: "front", Ipv6: net.ParseIP("fd00::5")}, "front fd00::5"},
		{"dual", v3.InstanceVpcSubnets{
			Name: "front",
			Ipv4: net.ParseIP("10.0.0.5"),
			Ipv6: net.ParseIP("fd00::5"),
		}, "front 10.0.0.5 fd00::5"},
		{"no address", v3.InstanceVpcSubnets{Name: "front"}, "front"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatInstanceVPCSubnet(tt.subnet); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
