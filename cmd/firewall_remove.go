package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var firewallRemoveCmd = &cobra.Command{
	Use:   "remove <security group name | id> <rule id | default rule name> [flags]\n  exo firewall remove <security group name | id> [flags]",
	Short: "Remove a rule from a security group",
}

func firewallRemoveRun(cmd *cobra.Command, args []string) {

	deleteAll, err := cmd.Flags().GetBool("all")
	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 1 && deleteAll {
		res, err := removeAllRules(args[0])

		for _, r := range res {
			println(r)
		}
		if err != nil {
			log.Fatal(err)
		}
		return
	}

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
		myCidr, err = getMyCIDR(isIpv6)
		if err != nil {
			log.Fatal(err)
		}
	}

	r, err := getDefaultRule(args[1], isIpv6)
	if err == nil {
		ru := &egoscale.IngressRule{
			Cidr:      r.CidrList[0],
			StartPort: r.StartPort,
			EndPort:   r.EndPort,
			Protocol:  r.Protocol,
		}
		removeDefault(args[0], args[1], ru, myCidr, isIpv6)
		return
	}

	removeRule(args[0], args[1])
}

func removeAllRules(sgName string) ([]string, error) {
	securGrp, err := getSecuGrpWithNameOrID(cs, sgName)
	if err != nil {
		return nil, err
	}

	res := []string{}

	for _, in := range securGrp.IngressRule {
		if err := cs.BooleanRequest(&egoscale.RevokeSecurityGroupIngress{ID: in.RuleID}); err != nil {
			return res, err
		}
		res = append(res, in.RuleID)
	}
	for _, eg := range securGrp.EgressRule {
		if err = cs.BooleanRequest(&egoscale.RevokeSecurityGroupEgress{ID: eg.RuleID}); err != nil {
			return res, err
		}
		res = append(res, eg.RuleID)
	}
	return res, nil
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

func init() {
	firewallRemoveCmd.Run = firewallRemoveRun
	firewallRemoveCmd.Flags().BoolP("ipv6", "6", false, "Remove rule with any IPv6 source")
	firewallRemoveCmd.Flags().BoolP("my-ip", "m", false, "Remove rule with my IP as a source")
	firewallRemoveCmd.Flags().BoolP("all", "", false, "Remove all rules")
	firewallCmd.AddCommand(firewallRemoveCmd)
}
