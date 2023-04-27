package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var (
	defaultCIDR  = egoscale.MustParseCIDR("0.0.0.0/0")
	defaultCIDR6 = egoscale.MustParseCIDR("::/0")
)

var firewallCmd = &cobra.Command{
	Use:   "firewall",
	Short: "Security Groups management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo firewall" commands are deprecated and will be removed in a future
version, please use "exo compute security-group" replacement commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
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
		if rule.IcmpCode == -1 || rule.IcmpType == -1 {
			desc = "Any"
		}

		ports = fmt.Sprintf("%d,%d (%s)", rule.IcmpType, rule.IcmpCode, desc)
	} else if rule.StartPort == rule.EndPort {
		ports = fmt.Sprint(rule.StartPort)
	} else {
		ports = fmt.Sprintf("%d-%d", rule.StartPort, rule.EndPort)
	}

	return ports
}

func getSecurityGroupByNameOrID(v string) (*egoscale.SecurityGroup, error) {
	sg := &egoscale.SecurityGroup{}

	id, err := egoscale.ParseUUID(v)

	if err != nil {
		sg.Name = v
	} else {
		sg.ID = id
	}

	resp, err := globalstate.EgoscaleClient.GetWithContext(gContext, sg)
	switch err {
	case nil:
		return resp.(*egoscale.SecurityGroup), nil

	case egoscale.ErrNotFound:
		return nil, fmt.Errorf("unknown Security Group %q", v)

	case egoscale.ErrTooManyFound:
		return nil, fmt.Errorf("multiple Security Groups match %q", v)

	default:
		return nil, err
	}
}

func getSecurityGroupIDs(params []string) ([]egoscale.UUID, error) {
	ids := make([]egoscale.UUID, len(params))

	for i, sg := range params {
		s, err := getSecurityGroupByNameOrID(sg)
		if err != nil {
			return nil, err
		}

		ids[i] = *s.ID
	}

	return ids, nil
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
