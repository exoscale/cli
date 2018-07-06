package cmd

import (
	"context"
	"fmt"
	"net"
	"regexp"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// firewallCmd represents the firewalling command
var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Security groups management",
}

const (
	defaultCidr  = "0.0.0.0/0"
	defaultCidr6 = "::/0"
)

func formatRules(name string, rule *egoscale.IngressRule) []string {
	var source string
	if rule.Cidr != "" {
		source = "CIDR " + rule.Cidr
	} else {
		source = "SG " + rule.SecurityGroupName
	}

	var ports string
	if rule.Protocol == "icmp" || rule.Protocol == "icmpv6" {
		c := icmpCode((uint16(rule.IcmpType) << 8) | uint16(rule.IcmpCode))
		t := c.icmpType()

		desc := c.StringFormatted()
		if desc == "" {
			desc = t.StringFormatted()
		}
		ports = fmt.Sprintf("%d, %d (%s)", rule.IcmpType, rule.IcmpCode, desc)
	} else if rule.StartPort == rule.EndPort {
		p := port(rule.StartPort)
		if p.StringFormatted() != "" {
			ports = fmt.Sprintf("%d (%s)", rule.StartPort, p.String())
		} else {
			ports = fmt.Sprintf("%d", rule.StartPort)
		}
	} else {
		ports = fmt.Sprintf("%d-%d", rule.StartPort, rule.EndPort)
	}

	return []string{name, source, rule.Protocol, ports, rule.Description, rule.RuleID}
}

func getSecuGrpWithNameOrID(cs *egoscale.Client, name string) (*egoscale.SecurityGroup, error) {
	if !isAFirewallID(cs, name) {
		securGrp := &egoscale.SecurityGroup{Name: name}
		if err := cs.Get(securGrp); err != nil {
			return nil, err
		}
		return securGrp, nil
	}

	securGrp := &egoscale.SecurityGroup{ID: name}
	if err := cs.Get(securGrp); err != nil {
		return nil, err
	}
	return securGrp, nil

}

func getMyCIDR(isIpv6 bool) (*net.IPNet, error) {

	var cidrMask net.IPMask
	dnsServer := ""

	if isIpv6 {
		dnsServer = "resolver2.ipv6-sandbox.opendns.com"
		cidrMask = net.CIDRMask(128, 128)
	} else {
		dnsServer = "resolver1.opendns.com"
		cidrMask = net.CIDRMask(32, 32)
	}
	resolver := net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial("udp", dnsServer+":53")
		},
		PreferGo: true,
	}

	ip, err := resolver.LookupIPAddr(context.Background(), "myip.opendns.com")
	if err != nil {
		return nil, err
	}

	if len(ip) < 1 {
		return nil, fmt.Errorf("no IP addresses were found using OpenDNS")
	}

	return &net.IPNet{IP: ip[0].IP, Mask: cidrMask}, nil
}

func isAFirewallID(cs *egoscale.Client, id string) bool {
	if !isUUID(id) {
		return false
	}
	req := &egoscale.SecurityGroup{ID: id}
	return cs.Get(req) == nil
}

// isUUID matches a UUIDv4
func isUUID(uuid string) bool {
	re := regexp.MustCompile(`(?i)^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
	return re.MatchString(uuid)
}

func init() {
	RootCmd.AddCommand(firewallCmd)
}
