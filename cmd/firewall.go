package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var defaultCIDR = egoscale.MustParseCIDR("0.0.0.0/0")
var defaultCIDR6 = egoscale.MustParseCIDR("::/0")

var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Security groups management",
}

func init() {
	RootCmd.AddCommand(firewallCmd)
}

func formatRuleSource(rule egoscale.IngressRule) string {
	var source string

	if rule.CIDR != nil {
		source = fmt.Sprintf("CIDR %s", rule.CIDR)
	} else {
		source = fmt.Sprintf("SG %s", rule.SecurityGroupName)
	}

	return source
}

func formatRulePort(rule egoscale.IngressRule) string {
	var ports string

	if rule.Protocol == "icmp" || rule.Protocol == "icmpv6" {
		c := icmpCode((uint16(rule.IcmpType) << 8) | uint16(rule.IcmpCode))
		t := c.icmpType()

		desc := c.StringFormatted()
		if desc == "" {
			desc = t.StringFormatted()
		}
		ports = fmt.Sprintf("%d,%d (%s)", rule.IcmpType, rule.IcmpCode, desc)
	} else if rule.StartPort == rule.EndPort {
		ports = fmt.Sprint(rule.StartPort)
	} else {
		ports = fmt.Sprintf("%d-%d", rule.StartPort, rule.EndPort)
	}

	return ports
}

func getSecurityGroupByNameOrID(name string) (*egoscale.SecurityGroup, error) {
	sg := &egoscale.SecurityGroup{}

	id, err := egoscale.ParseUUID(name)

	if err != nil {
		sg.Name = name
	} else {
		sg.ID = id
	}

	resp, err := cs.GetWithContext(gContext, sg)
	if err != nil {
		return nil, err
	}
	return resp.(*egoscale.SecurityGroup), nil

}

func getMyCIDR(isIpv6 bool) (*egoscale.CIDR, error) {
	cidrMask := 32
	dnsServer := "resolver1.opendns.com"
	protocol := "udp4"

	if isIpv6 {
		dnsServer = "resolver2.ipv6-sandbox.opendns.com"
		cidrMask = 128
		protocol = "udp6"
	}

	resolver := net.Resolver{
		Dial: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial(protocol, dnsServer+":53")
		},
		PreferGo: true,
	}

	ips, err := resolver.LookupIPAddr(gContext, "myip.opendns.com")
	if err != nil {
		return nil, err
	}

	if len(ips) < 1 {
		return nil, fmt.Errorf("no IP addresses were found using OpenDNS")
	}

	return egoscale.ParseCIDR(fmt.Sprintf("%s/%d", ips[0].IP, cidrMask))
}
