package egoscale

import (
	"encoding/json"
	"fmt"
	"net"
)

// CIDR represents a nicely JSON serializable net.IPNet
type CIDR struct {
	*net.IPNet
}

// UnmarshalJSON unmarshals the raw JSON into the MAC address
func (cidr *CIDR) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	c, err := ParseCIDR(s)
	if err != nil {
		return err
	}
	cidr.IPNet = c.IPNet
	return nil
}

// MarshalJSON converts the CIDR to a string representation
func (cidr CIDR) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", cidr.IPNet)), nil
}

// ParseCIDR parses a CIDR from a string
func ParseCIDR(s string) (*CIDR, error) {
	ip, net, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	if net == nil || !ip.Equal(net.IP) {
		return nil, fmt.Errorf("missmatch between CIDR address %q and network number %q", net.IP, ip)
	}

	return &CIDR{net}, nil
}

// ForceParseCIDR forces parseCIDR or panics
func ForceParseCIDR(s string) *CIDR {
	cidr, err := ParseCIDR(s)
	if err != nil {
		panic(err)
	}

	return cidr
}
