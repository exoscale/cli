package cmd

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// firewallAddCmd represents the add command
var firewallAddCmd = &cobra.Command{
	Use:   "add [<security group name> | <id>] [ssh | telnet | rdp | ...]",
	Short: "Add rule to a security group",
}

func firewallAddRun(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		firewallAddCmd.Usage()
		return
	}

	isMyIP, err := cmd.Flags().GetBool("my-ip")
	if err != nil {
		log.Fatal(err)
	}

	isIpv6, err := cmd.Flags().GetBool("ipv6")
	if err != nil {
		log.Fatal(err)
	}

	ip := ""
	if isMyIP {
		ip = getMyCIDR(isIpv6)

	}

	if isIpv6 && ip == "" {
		ip = "::/0"
	}

	desc, err := cmd.Flags().GetString("description")
	if err != nil {
		log.Fatal(err)
	}

	addDefaultRule(args[0], args[1], ip, desc)
}

func addDefaultRule(sg, ruleName, cidr, desc string) {
	rule, ok := defaultRules[strings.ToLower(ruleName)]
	if !ok {
		log.Fatalf("Rule: '%s' not found or doesn't exist\n", ruleName)
	}

	securGrp, err := getSecuGrpWithNameOrID(cs, sg)
	if err != nil {
		log.Fatal(err)
	}

	if cidr != "" {
		rule.Cidr = cidr
	}

	req := &egoscale.AuthorizeSecurityGroupIngress{
		Protocol:        rule.Protocol,
		CidrList:        []string{rule.Cidr},
		SecurityGroupID: securGrp.ID,
		Description:     rule.Description,
	}

	if rule.Protocol == "icmp" || rule.Protocol == "icmpv6" {
		req.IcmpType = rule.IcmpType
		req.IcmpCode = rule.IcmpCode
	} else {
		req.StartPort = rule.StartPort
		req.EndPort = rule.EndPort
	}

	_, err = cs.Request(req)
	if err != nil {
		log.Fatal(err)
	}
	firewallDetails(sg)
}

func getMyCIDR(isIpv6 bool) string {

	cidrSuffix := ""
	dnsServer := ""

	if isIpv6 {
		dnsServer = "resolver1.ipv6-sandbox.opendns.com"
		cidrSuffix = "/128"
	} else {
		dnsServer = "resolver1.opendns.com"
		cidrSuffix = "/32"
	}
	resolver := net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial("udp", dnsServer+":53")
		},
		PreferGo: true,
	}

	ip, err := resolver.LookupIPAddr(context.Background(), "myip.opendns.com")
	if err != nil {
		log.Fatal(err)
	}

	if len(ip) < 1 {
		return ""
	}

	return (ip[0].IP.String() + cidrSuffix)

}

func init() {
	firewallAddCmd.Run = firewallAddRun

	firewallAddCmd.Flags().BoolP("ipv6", "6", false, "Add rule for any IPv6 source")
	firewallAddCmd.Flags().BoolP("my-ip", "m", false, "Add rule only for my IP as a source")

	// firewallAddCmd.Flags().StringP("type", "t", "", "Rule type available [INGRESS or EGRESS]")
	// firewallAddCmd.Flags().StringP("protocol", "p", "", "Rule Protocol available [tcp, udp, icmp, icmpv6, ah, esp, gre]")
	// firewallAddCmd.Flags().StringP("source", "s", "", "Rule Source [CIDR 0.0.0.0/0,192.168.0.0/16 or security group name or id]")
	// firewallAddCmd.Flags().StringP("port", "P", "", "Rule port range [80-80,443,22-22]")
	firewallAddCmd.Flags().StringP("description", "d", "", "Rule description")
	firewallCmd.AddCommand(firewallAddCmd)
}
