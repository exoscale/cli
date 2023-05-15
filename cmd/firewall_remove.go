package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/egoscale"
)

var firewallRemoveCmd = &cobra.Command{
	Use:     "remove SECURITY-GROUP-NAME|ID RULE-ID|DEFAULT-RULE-NAME",
	Short:   "Remove a rule from a Security Group",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		deleteAll, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		sg, errGet := getSecurityGroupByNameOrID(args[0])
		if errGet != nil {
			return errGet
		}

		if len(args) == 1 && deleteAll {
			count := len(sg.IngressRule) + len(sg.EgressRule)
			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete all %d firewall rule(s) from %s", count, sg.Name)) {
					return nil
				}
			}
			return removeAllRules(sg)
		}

		if len(args) < 2 {
			return cmd.Usage()
		}

		isMyIP, err := cmd.Flags().GetBool("my-ip")
		if err != nil {
			return err
		}

		isIpv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		cidr := defaultCIDR
		if isIpv6 {
			cidr = defaultCIDR6
		}
		if isMyIP {
			c, errGet := getMyCIDR(isIpv6)
			if errGet != nil {
				return errGet
			}
			cidr = c
		}

		tasks := make([]task, 0, len(args[1:]))

		for _, arg := range args[1:] {
			var ruleID *egoscale.UUID

			if !force {
				if !askQuestion(fmt.Sprintf("Are your sure you want to delete the %q firewall rule from %s", arg, sg.Name)) {
					continue
				}
			}

			r, err := getDefaultRule(arg)
			if err == nil {
				ru := &egoscale.IngressRule{
					CIDR:      cidr,
					StartPort: r.StartPort,
					EndPort:   r.EndPort,
					Protocol:  r.Protocol,
				}
				ruleID, err = prepareDefaultRemove(sg, arg, ru, cidr, isIpv6)
				if err != nil {
					return err
				}
				tasks = append(tasks, task{egoscale.RevokeSecurityGroupIngress{ID: ruleID}, fmt.Sprintf("Remove %q rule", arg)})
				continue
			}
			err = removeRule(sg, arg, &tasks)
			if err != nil {
				return err
			}
		}
		resps := asyncTasks(tasks)
		errs := filterErrors(resps)
		if len(errs) > 0 {
			return errs[0]
		}

		return nil
	},
}

func removeAllRules(sg *egoscale.SecurityGroup) error {
	tasks := []task{}

	for _, in := range sg.IngressRule {
		tasks = append(tasks, task{&egoscale.RevokeSecurityGroupIngress{ID: in.RuleID}, fmt.Sprintf("Remove %q rule", in.RuleID)})
	}
	for _, eg := range sg.EgressRule {
		tasks = append(tasks, task{&egoscale.RevokeSecurityGroupEgress{ID: eg.RuleID}, fmt.Sprintf("Remove %q rule", eg.RuleID)})
	}

	resps := asyncTasks(tasks)
	errs := filterErrors(resps)
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func removeRule(sg *egoscale.SecurityGroup, ruleID string, tasks *[]task) error {
	id, err := egoscale.ParseUUID(ruleID)
	if err != nil {
		return err
	}

	in, eg := sg.RuleByID(*id)

	var msg string
	if in != nil { //nolint:gocritic
		msg = fmt.Sprintf("Remove %q", in.RuleID)
		*tasks = append(*tasks, task{egoscale.RevokeSecurityGroupIngress{ID: in.RuleID}, msg})
	} else if eg != nil {
		msg = fmt.Sprintf("Remove %q", eg.RuleID)
		*tasks = append(*tasks, task{egoscale.RevokeSecurityGroupEgress{ID: eg.RuleID}, msg})
	} else {
		return fmt.Errorf("rule with id %q doesn't exist", ruleID)
	}
	return nil
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
		rule.CIDR.Equal(*cidr) &&
		rule.Protocol == defaultRule.Protocol)
}

func prepareDefaultRemove(sg *egoscale.SecurityGroup, ruleName string, rule *egoscale.IngressRule, cidr *egoscale.CIDR, isIpv6 bool) (*egoscale.UUID, error) {
	for i, in := range sg.IngressRule {
		if !isDefaultRule(&sg.IngressRule[i], rule, isIpv6, cidr) {
			// Rule not found
			continue
		}
		return in.RuleID, nil
	}
	return nil, fmt.Errorf("missing rule %q", ruleName)
}

func init() {
	firewallRemoveCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	firewallRemoveCmd.Flags().BoolP("ipv6", "6", false, "Remove rule with any IPv6 source")
	firewallRemoveCmd.Flags().BoolP("my-ip", "m", false, "Remove rule with my IP as a source")
	firewallRemoveCmd.Flags().BoolP("all", "a", false, "Remove all rules")
	firewallCmd.AddCommand(firewallRemoveCmd)
}
