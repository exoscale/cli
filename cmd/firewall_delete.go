package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	firewallDeleteCmd.Flags().BoolP("force", "f", false, "Delete security group without prompting for confirmation")
	firewallDeleteCmd.Flags().BoolP("all", "", false, "Delete all security group and rules")
	firewallCmd.AddCommand(firewallDeleteCmd)
}

// deleteCmd represents the delete command
var firewallDeleteCmd = &cobra.Command{
	Use:     "delete <security group name | id>+",
	Short:   "Delete security group",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if len(args) < 1 && !all {
			return cmd.Usage()
		}

		if all {
			q := fmt.Sprintf("Are you sure you want to delete all security groups")
			if !force && !askQuestion(q) {
				return nil
			}
			r, err := cs.ListWithContext(gContext, &egoscale.SecurityGroup{})
			if err != nil {
				return err
			}
			args = make([]string, 0, len(r))
			for _, s := range r {
				sg := s.(*egoscale.SecurityGroup)
				if sg.Name == "default" {
					if err := removeAllRules(sg); err != nil {
						return err
					}
					continue
				}
				args = append(args, sg.ID.String())
			}
		}

		sgTasks := make([]task, 0, len(args))
		var rulesTask []task
		for _, arg := range args {
			sg, err := getSecurityGroupByNameOrID(arg)
			if err != nil {
				return err
			}

			q := fmt.Sprintf("Are you sure you want to delete the security group: %q", sg.Name)
			if !all && !force && !askQuestion(q) {
				continue
			}

			for _, r := range sg.IngressRule {
				rulesTask = append(rulesTask, task{
					&egoscale.RevokeSecurityGroupIngress{
						ID: r.RuleID,
					},
					fmt.Sprintf("deleting %q rule from %q", r.RuleID, sg.Name),
				})
			}
			for _, r := range sg.EgressRule {
				rulesTask = append(rulesTask, task{
					&egoscale.RevokeSecurityGroupEgress{
						ID: r.RuleID,
					},
					fmt.Sprintf("deleting %q rule from %q", r.RuleID, sg.Name),
				})
			}

			cmd := &egoscale.DeleteSecurityGroup{ID: sg.ID}
			sgTasks = append(sgTasks, task{
				cmd,
				fmt.Sprintf("delete %q SG", sg.Name),
			})
		}

		if len(rulesTask) > 0 {
			ruleResps := asyncTasks(rulesTask)
			errs := filterErrors(ruleResps)
			if len(errs) > 0 {
				return errs[0]
			}
		}

		sgResps := asyncTasks(sgTasks)
		errs := filterErrors(sgResps)
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	},
}
