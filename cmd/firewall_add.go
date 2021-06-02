package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type portRange struct {
	start uint16
	end   uint16
}

func init() {
	firewallAddCmd.Flags().BoolP("ipv6", "6", false, "Set ipv6 on default rules or on --my-ip")
	firewallAddCmd.Flags().BoolP("my-ip", "", false, "Set CIDR for my ip")
	firewallAddCmd.Flags().BoolP("egress", "e", false, "By default rule is INGRESS (set --egress to have EGRESS rule)")
	firewallAddCmd.Flags().StringP("protocol", "p", "", "Rule Protocol available [tcp, udp, icmp, icmpv6, ah, esp, gre]")
	firewallAddCmd.Flags().StringP("cidr", "c", "", "Rule CIDR [CIDR 0.0.0.0/0,::/0,...]")
	firewallAddCmd.Flags().StringP("security-group", "s", "", "Rule Security Group [NAME|ID ex: sg1,sg2...]")
	firewallAddCmd.Flags().StringP("port", "P", "", "Rule port range [80-80,443,22-22]")
	firewallAddCmd.Flags().Int("icmp-type", -1, "Rule ICMP type")
	firewallAddCmd.Flags().Int("icmp-code", -1, "Rule ICMP code")
	firewallAddCmd.Flags().StringP("description", "d", "", "Rule description")

	firewallCmd.AddCommand(firewallAddCmd)
}

var firewallAddCmd = &cobra.Command{
	Use:   "add SECURITY-GROUP-NAME|ID [ping|ssh|rdp]",
	Short: "Add a rule to a Security Group",
	Long: `
No arguments create a IPv4 wide open rule.

For ICMP rules, specify the icmp-code and icmp-type.

	firewall add <Security Group> --protocol icmp --icmp-type 8 --icmp-code 0

For the other protocols specify the port ranges.

	firewall add <Security Group> --protocol tcp --port 8000-8080

A set of predefined commands exists: ping, ssh, rdp.

	firewall add <Security Group> ssh
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		securityGroup, err := getSecurityGroupByNameOrID(args[0])
		if err != nil {
			return err
		}

		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}

		protocol, err := cmd.Flags().GetString("protocol")
		if err != nil {
			return err
		}

		isEgress, err := cmd.Flags().GetBool("egress")
		if err != nil {
			return err
		}

		icmptype, err := cmd.Flags().GetInt("icmp-type")
		if err != nil {
			return err
		}

		icmpcode, err := cmd.Flags().GetInt("icmp-code")
		if err != nil {
			return err
		}

		port, err := cmd.Flags().GetString("port")
		if err != nil {
			return err
		}

		cidrList, err := cmd.Flags().GetString("cidr")
		if err != nil {
			return err
		}

		sg, err := cmd.Flags().GetString("security-group")
		if err != nil {
			return err
		}

		isIpv6, err := cmd.Flags().GetBool("ipv6")
		if err != nil {
			return err
		}

		isMyIP, err := cmd.Flags().GetBool("my-ip")
		if err != nil {
			return err
		}

		if port != "" && protocol == "" {
			return errors.New(`"--port" can only be specified with "--protocol"`)
		}

		var ip *egoscale.CIDR
		if isMyIP {
			cidr, cirdErr := getMyCIDR(isIpv6)
			if cirdErr != nil {
				return cirdErr
			}

			ip = cidr
		}

		tasks := []task{}

		for i := 1; true; i++ {
			if i >= len(args) && len(args) != 1 {
				break
			}

			rule := &egoscale.AuthorizeSecurityGroupIngress{}

			if len(args) > 1 {
				rule, err = getDefaultRule(args[i])
				if err != nil {
					return err
				}
			}

			rule.Description = desc
			rule.SecurityGroupID = securityGroup.ID

			if protocol != "" {
				rule.Protocol = strings.ToLower(protocol)
			}

			if ip != nil {
				rule.CIDRList = append(rule.CIDRList, *ip)
			}

			if cidrList != "" {
				cidrs := getCommaflag(cidrList)
				for _, cidr := range cidrs {
					c, errCidr := egoscale.ParseCIDR(cidr)
					if errCidr != nil {
						return errCidr
					}
					rule.CIDRList = append(rule.CIDRList, *c)
				}
			}

			if sg != "" {
				sgs := getCommaflag(sg)

				userSecurityGroups, sgErr := getUserSecurityGroups(sgs)
				if sgErr != nil {
					return sgErr
				}

				rule.UserSecurityGroupList = userSecurityGroups
			} else if len(rule.CIDRList) == 0 {
				if (isIpv6 || rule.Protocol == "ICMPv6") && rule.Protocol != "ICMP" {
					rule.CIDRList = append(rule.CIDRList, *defaultCIDR6)
				} else {
					rule.CIDRList = append(rule.CIDRList, *defaultCIDR)
				}
			}

			if strings.HasPrefix(rule.Protocol, "icmp") {
				rule.IcmpType = icmptype
				rule.IcmpCode = icmpcode
			}

			// Not best practice but waiting to find better solution
			if port != "" && (rule.Protocol == "tcp" || rule.Protocol == "udp") {
				ports := getCommaflag(port)
				portsRange, err := getPortsRange(ports)
				if err != nil {
					return err
				}

				for _, portRange := range portsRange {
					rule.StartPort = portRange.start
					rule.EndPort = portRange.end

					msg := fmt.Sprintf("Add rule for %q with port %d", securityGroup.Name, rule.StartPort)
					tasks = append(tasks, newFirewallRuleTask(*rule, msg, isEgress))
				}
			}

			// Not best practice but waiting to find better solution
			if port == "" || !(rule.Protocol == "tcp" || rule.Protocol == "udp") {
				msg := fmt.Sprintf("Add rule for %q", securityGroup.Name)
				if len(args) > 1 {
					msg = fmt.Sprintf("Add %q rule for %q", args[i], securityGroup.Name)
				}
				tasks = append(tasks, newFirewallRuleTask(*rule, msg, isEgress))
			}

			if len(args) == 1 {
				break
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

func getPortsRange(ports []string) ([]portRange, error) {
	portsRange := make([]portRange, len(ports))
	for i, p := range ports {
		pRange := strings.Split(p, "-")
		if len(pRange) > 2 || len(pRange) == 0 {
			return nil, fmt.Errorf("failed to find port ranges into: %q", p)
		}
		p1, err := strconv.ParseUint(pRange[0], 10, 16)
		if err != nil {
			return nil, err
		}

		portsRange[i].start = uint16(p1)
		portsRange[i].end = uint16(p1)

		if len(pRange) == 2 {
			p2, err := strconv.ParseUint(pRange[1], 10, 16)
			if err != nil {
				return nil, err
			}
			portsRange[i].end = uint16(p2)
		}
	}
	return portsRange, nil
}

func getUserSecurityGroups(names []string) ([]egoscale.UserSecurityGroup, error) {
	us := make([]egoscale.UserSecurityGroup, 0, len(names))
	for _, sg := range names {
		s, err := getSecurityGroupByNameOrID(sg)
		if err != nil {
			return nil, err
		}

		us = append(us, s.UserSecurityGroup())
	}
	return us, nil
}

func getDefaultRule(ruleName string) (*egoscale.AuthorizeSecurityGroupIngress, error) {
	ruleName = strings.ToLower(ruleName)
	if ruleName == "ping" {
		return &egoscale.AuthorizeSecurityGroupIngress{
			Protocol: "ICMP",
			CIDRList: []egoscale.CIDR{},
			IcmpType: int(echo),
			IcmpCode: 0,
		}, nil
	}

	if ruleName == "ping6" {
		return &egoscale.AuthorizeSecurityGroupIngress{
			Protocol: "ICMPv6",
			CIDRList: []egoscale.CIDR{},
			IcmpType: int(echoRequest),
			IcmpCode: 0,
		}, nil
	}

	for d := Daytime; d <= Minecraft; d++ {
		if strings.ToLower(port.String(d)) == ruleName {
			return &egoscale.AuthorizeSecurityGroupIngress{
				Protocol:  "TCP",
				CIDRList:  []egoscale.CIDR{},
				StartPort: uint16(d),
				EndPort:   uint16(d),
			}, nil
		}
	}

	return nil, fmt.Errorf("default rule %q not found", ruleName)
}

func newFirewallRuleTask(rule egoscale.AuthorizeSecurityGroupIngress, msg string, isEgress bool) task {
	if isEgress {
		req := (egoscale.AuthorizeSecurityGroupEgress)(rule)
		return task{req, msg}
	}

	return task{rule, msg}
}
