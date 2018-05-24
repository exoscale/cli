package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var firewallRemoveCmd = &cobra.Command{
	Use:   "remove <security group name | id> <rule id | default rule name>",
	Short: "Remove a rule from a security group",
}

func firewallRemoveRun(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		firewallRemoveCmd.Usage()
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

	var myCidr *net.IPNet
	if isMyIP {
		myCidr, err = _getMyCIDR(isIpv6)
		if err != nil {
			log.Fatal(err)
		}
	}

	r, ok := defaultRules[strings.ToLower(args[1])]
	if ok {
		removeDefault(args[0], args[1], r, myCidr, isIpv6)
		return
	}

	removeRule(args[0], args[1])
}

func removeRule(sg, ruleID string) {
	securGrp, err := getSecuGrpWithNameOrID(cs, sg)
	if err != nil {
		log.Fatal(err)
	}

	in, eg := securGrp.RuleByID(ruleID)

	if in != nil {
		err = cs.BooleanRequest(&egoscale.RevokeSecurityGroupIngress{ID: in.RuleID})
	} else if eg != nil {
		err = cs.BooleanRequest(&egoscale.RevokeSecurityGroupEgress{ID: eg.RuleID})
	} else {
		err = fmt.Errorf("rule with id %q is not ingress or egress rule", ruleID)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ruleID)
}

func isDefaultRule(rule, defaultRule *egoscale.IngressRule, isIpv6 bool, myCidr *net.IPNet) bool {
	cidr := "0.0.0.0/0"
	if isIpv6 {
		cidr = "::/0"
	}

	if myCidr != nil {
		cidr = myCidr.String()
	}

	return (rule.StartPort == defaultRule.StartPort &&
		rule.EndPort == defaultRule.EndPort &&
		rule.Cidr == cidr &&
		rule.Protocol == defaultRule.Protocol)
}

func removeDefault(sgName, ruleName string, rule *egoscale.IngressRule, cidr *net.IPNet, isIpv6 bool) {
	securGrp, err := getSecuGrpWithNameOrID(cs, sgName)
	if err != nil {
		log.Fatal(err)
	}

	for _, in := range securGrp.IngressRule {
		if isDefaultRule(&in, rule, isIpv6, nil) && cidr == nil {
			//Rule found
		} else if isDefaultRule(&in, rule, isIpv6, cidr) {
			//Rule found
		} else {
			//Rule not found
			continue
		}
		err := cs.BooleanRequest(&egoscale.RevokeSecurityGroupIngress{ID: in.RuleID})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(in.RuleID)
		return
	}
	log.Fatalf("Rule %q not foud", ruleName)
}

//Waiting to complete-firewall-rule branch merge
func _getMyCIDR(isIpv6 bool) (*net.IPNet, error) {

	var cidrMask net.IPMask
	dnsServer := ""

	if isIpv6 {
		dnsServer = "resolver1.ipv6-sandbox.opendns.com"
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
		return nil, fmt.Errorf("Invalide IP adress")
	}

	return &net.IPNet{IP: ip[0].IP, Mask: cidrMask}, nil
}

func init() {
	firewallRemoveCmd.Run = firewallRemoveRun
	firewallRemoveCmd.Flags().BoolP("ipv6", "6", false, "Remove rule with any IPv6 source")
	firewallRemoveCmd.Flags().BoolP("my-ip", "m", false, "Remove rule with my IP as a source")
	firewallCmd.AddCommand(firewallRemoveCmd)
}
