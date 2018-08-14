package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	firewallRemoveCmd.Flags().BoolP("force", "f", false, "Attempt to remove firewall rule without prompting for confirmation")
	firewallRemoveCmd.Flags().BoolP("ipv6", "6", false, "Remove rule with any IPv6 source")
	firewallRemoveCmd.Flags().BoolP("my-ip", "m", false, "Remove rule with my IP as a source")
	firewallRemoveCmd.Flags().BoolP("all", "", false, "Remove all rules")
	firewallCmd.AddCommand(firewallRemoveCmd)
}

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

		sgName := args[0]

		if len(args) == 1 && deleteAll {
			sg, errGet := getSecurityGroupByNameOrID(sgName)
			if errGet != nil {
				return errGet
			}
			count := len(sg.IngressRule) + len(sg.EgressRule)
			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete all %d firewall rule(s) from %s", count, sgName)) {
					return nil
				}
			}
			res, rErr := removeAllRules(sgName)

			for _, r := range res {
				println(r)
			}
			return rErr
		}

		if len(args) < 2 {
			return cmd.Usage()
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are your sure you want to delete the %q firewall rule from %s", args[1], sgName)) {
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

		var cidr *egoscale.CIDR
		if isMyIP {
			c, errGet := getMyCIDR(isIpv6)
			if errGet != nil {
				return errGet
			}
			cidr = c
		}

		r, err := getDefaultRule(args[1], isIpv6)
		if err == nil {
			ru := &egoscale.IngressRule{
				CIDR:      &r.CIDRList[0],
				StartPort: r.StartPort,
				EndPort:   r.EndPort,
				Protocol:  r.Protocol,
			}
			return removeDefault(args[0], args[1], ru, cidr, isIpv6)
		}

		return removeRule(args[0], args[1])
	},
}

func removeAllRules(sgName string) ([]string, error) {
	sg, err := getSecurityGroupByNameOrID(sgName)
	if err != nil {
		return nil, err
	}

	res := []string{}

	for _, in := range sg.IngressRule {
		if reqErr := cs.BooleanRequestWithContext(gContext, &egoscale.RevokeSecurityGroupIngress{ID: in.RuleID}); reqErr != nil {
			return res, reqErr
		}
		res = append(res, in.RuleID.String())
	}
	for _, eg := range sg.EgressRule {
		if err = cs.BooleanRequestWithContext(gContext, &egoscale.RevokeSecurityGroupEgress{ID: eg.RuleID}); err != nil {
			return res, err
		}
		res = append(res, eg.RuleID.String())
	}
	return res, nil
}

func removeRule(name, ruleID string) error {
	sg, err := getSecurityGroupByNameOrID(name)
	if err != nil {
		return err
	}

	id, err := egoscale.ParseUUID(ruleID)
	if err != nil {
		return err
	}

	in, eg := sg.RuleByID(*id)

	if in != nil {
		err = cs.BooleanRequestWithContext(gContext, &egoscale.RevokeSecurityGroupIngress{ID: in.RuleID})
	} else if eg != nil {
		err = cs.BooleanRequestWithContext(gContext, &egoscale.RevokeSecurityGroupEgress{ID: eg.RuleID})
	} else {
		err = fmt.Errorf("rule with id %q is not ingress or egress rule", ruleID)
	}

	if err != nil {
		return err
	}

	_, err = fmt.Println(ruleID)
	return err
}

func isDefaultRule(rule, defaultRule *egoscale.IngressRule, isIpv6 bool, myCidr *egoscale.CIDR) bool {
	cidr := defaultCIDR
	if isIpv6 {
		cidr = defaultCIDR6
	}

	if myCidr != nil {
		cidr = myCidr
	}

	return (rule.StartPort == defaultRule.StartPort &&
		rule.EndPort == defaultRule.EndPort &&
		rule.CIDR == cidr &&
		rule.Protocol == defaultRule.Protocol)
}

func removeDefault(sgName, ruleName string, rule *egoscale.IngressRule, cidr *egoscale.CIDR, isIpv6 bool) error {
	sg, err := getSecurityGroupByNameOrID(sgName)
	if err != nil {
		return err
	}

	for _, in := range sg.IngressRule {
		if !isDefaultRule(&in, rule, isIpv6, cidr) {
			// Rule not found
			continue
		}

		err := cs.BooleanRequestWithContext(gContext, &egoscale.RevokeSecurityGroupIngress{ID: in.RuleID})
		if err != nil {
			return err
		}

		fmt.Println(in.RuleID)
		return nil
	}
	return fmt.Errorf("missing rule %q", ruleName)
}
