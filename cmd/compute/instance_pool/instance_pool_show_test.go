package instance_pool

import (
	"testing"

	v3 "github.com/exoscale/egoscale/v3"
)

func TestFormatVPCSubnetBlocks(t *testing.T) {
	tests := []struct {
		name   string
		subnet v3.ListSubnetEntry
		want   string
	}{
		{"inet4", v3.ListSubnetEntry{Name: "front", Ipv4Block: "10.0.0.0/24"}, "front 10.0.0.0/24"},
		{"inet6", v3.ListSubnetEntry{Name: "front", Ipv6Block: "fd00:1::/64"}, "front fd00:1::/64"},
		{"dual", v3.ListSubnetEntry{
			Name:      "front",
			Ipv4Block: "10.0.0.0/24",
			Ipv6Block: "fd00:1::/64",
		}, "front 10.0.0.0/24 fd00:1::/64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatVPCSubnetBlocks(tt.subnet); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
