package cmd

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var firewallRemoveCmd = &cobra.Command{
	Use:     "remove <security group name | id> <rule id | default rule name> [flags]\n  exo firewall remove <security group name | id> [flags]",
	Short:   "Remove a rule from a security group",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {

		deleteAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if len(args) == 1 && deleteAll {
			if !force {
				if !askQuestion(fmt.Sprintf("sure you want to delete all %d firewall rules", len(args)-1)) {
					return nil
				}
			}
			res, rErr := removeAllRules(args[0])

			for _, r := range res {
				println(r)
			}
			return rErr
		}

		if len(args) < 2 {
			return cmd.Usage()
		}

		if !force {
			if !askQuestion(fmt.Sprintf("sure you want to delete %q firewall rule", args[0])) {
				return nil
			}
		}

		isMyIP, err := cmd.Flags().GetBool("my-ip")
		if err != nil {
			return err
		}

		isIpv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		var myCidr string
		var cidr *net.IPNet
		if isMyIP {
			cidr, err = getMyCIDR(isIpv6)
			if err != nil {
				return err
			}
			myCidr = cidr.String()
		}

		r, err := getDefaultRule(args[1], isIpv6)
		if err == nil {
			ru := &egoscale.IngressRule{
				Cidr:      r.CidrList[0],
				StartPort: r.StartPort,
				EndPort:   r.EndPort,
				Protocol:  r.Protocol,
			}
			return removeDefault(args[0], args[1], ru, myCidr, isIpv6)
		}

		return removeRule(args[0], args[1])
	},
}

func removeAllRules(sgName string) ([]string, error) {
	securGrp, err := getSecuGrpWithNameOrID(cs, sgName)
	if err != nil {
		return nil, err
	}

	res := []string{}

	for _, in := range securGrp.IngressRule {
		if reqErr := cs.BooleanRequest(&egoscale.RevokeSecurityGroupIngress{ID: in.RuleID}); reqErr != nil {
			return res, reqErr
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

func removeRule(sg, ruleID string) error {
	securGrp, err := getSecuGrpWithNameOrID(cs, sg)
	if err != nil {
		return err
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
		return err
	}

	_, err = fmt.Println(ruleID)
	return err
}

func isDefaultRule(rule, defaultRule *egoscale.IngressRule, isIpv6 bool, myCidr string) bool {
	cidr := "0.0.0.0/0"
	if isIpv6 {
		cidr = "::/0"
	}

	if myCidr != "" {
		cidr = myCidr
	}

	return (rule.StartPort == defaultRule.StartPort &&
		rule.EndPort == defaultRule.EndPort &&
		rule.Cidr == cidr &&
		rule.Protocol == defaultRule.Protocol)
}

func removeDefault(sgName, ruleName string, rule *egoscale.IngressRule, cidr string, isIpv6 bool) error {
	securGrp, err := getSecuGrpWithNameOrID(cs, sgName)
	if err != nil {
		return err
	}

	for _, in := range securGrp.IngressRule {
		if isDefaultRule(&in, rule, isIpv6, "") && cidr == "" {
			//Rule found
		} else if isDefaultRule(&in, rule, isIpv6, cidr) {
			//Rule found
		} else {
			//Rule not found
			continue
		}
		err := cs.BooleanRequest(&egoscale.RevokeSecurityGroupIngress{ID: in.RuleID})
		if err != nil {
			return err
		}
		_, err = fmt.Println(in.RuleID)
		return err
	}
	return fmt.Errorf("Rule %q not foud", ruleName)
}

func init() {
	firewallRemoveCmd.Flags().BoolP("force", "f", false, "Attempt to remove firewall rule without prompting for confirmation")
	firewallRemoveCmd.Flags().BoolP("ipv6", "6", false, "Remove rule with any IPv6 source")
	firewallRemoveCmd.Flags().BoolP("my-ip", "m", false, "Remove rule with my IP as a source")
	firewallRemoveCmd.Flags().BoolP("all", "", false, "Remove all rules")
	firewallCmd.AddCommand(firewallRemoveCmd)
}
