package cmd

import (
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
	firewallAddCmd.Flags().StringP("cidr", "c", "", "Rule Cidr [CIDR 0.0.0.0/0,::/0,...]")
	firewallAddCmd.Flags().StringP("security-group", "s", "", "Rule security group [name or id ex: sg1,sg2...]")
	firewallAddCmd.Flags().StringP("port", "P", "", "Rule port range [80-80,443,22-22]")

	//Flag for icmp
	icmpTypeVarP := new(uint8PtrValue)
	icmpCodeVarP := new(uint8PtrValue)

	firewallAddCmd.Flags().VarP(icmpTypeVarP, "icmp-type", "", "Set icmp type")
	firewallAddCmd.Flags().VarP(icmpCodeVarP, "icmp-code", "", "Set icmp type code")

	firewallAddCmd.Flags().StringP("description", "d", "", "Rule description")

	firewallCmd.AddCommand(firewallAddCmd)
}

// firewallAddCmd represents the add command
var firewallAddCmd = &cobra.Command{
	Use:   "add <security group name | id>  [ssh | telnet | rdp | ...] (default preset rules)",
	Short: "Add rule to a security group",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		securityGroup, err := getSecurityGroupByNameOrID(cs, args[0])
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

		icmptype, err := getUint8CustomFlag(cmd, "icmp-type")
		if err != nil {
			return err
		}

		icmpcode, err := getUint8CustomFlag(cmd, "icmp-code")
		if err != nil {
			return err
		}

		port, err := cmd.Flags().GetString("port")
		if err != nil {
			return err
		}

		cidr, err := cmd.Flags().GetString("cidr")
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

		ip := ""
		if isMyIP {
			cidr, cirdErr := getMyCIDR(isIpv6)
			if cirdErr != nil {
				return cirdErr
			}

			ip = cidr.String()
		}

		for i := 1; true; i++ {
			if i >= len(args) && len(args) != 1 {
				break
			}

			rule := &egoscale.AuthorizeSecurityGroupIngress{}

			if len(args) != 0 {
				rule, err = getDefaultRule(args[i], isIpv6)
				if err != nil {
					return err
				}
			}

			rule.Description = desc
			rule.SecurityGroupID = securityGroup.ID

			if protocol != "" {
				rule.Protocol = strings.ToLower(protocol)
			}

			if ip != "" {
				rule.CidrList = []string{ip}
			} else if cidr != "" {
				cidrs := getCommaflag(cidr)
				rule.CidrList = append(rule.CidrList, cidrs...)
			}

			if sg != "" {
				sgs := getCommaflag(sg)

				userSecurityGroups, sgErr := getUserSecurityGroups(cs, sgs)
				if sgErr != nil {
					return sgErr
				}

				rule.UserSecurityGroupList = userSecurityGroups
			}

			if icmptype.uint8 != nil {
				rule.IcmpType = *icmptype.uint8
			}

			if icmpcode.uint8 != nil {
				rule.IcmpCode = *icmpcode.uint8
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
					if err := addRule(rule, isEgress); err != nil {
						return err
					}
				}
			}

			// Not best practice but waiting to find better solution
			if port == "" || !(rule.Protocol == "tcp" || rule.Protocol == "udp") {
				if err := addRule(rule, isEgress); err != nil {
					return err
				}
			}
		}

		return firewallShow.RunE(cmd, []string{securityGroup.ID})
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
			p2, err := strconv.ParseUint(pRange[0], 10, 16)
			if err != nil {
				return nil, err
			}
			portsRange[i].end = uint16(p2)
		}
	}
	return portsRange, nil
}

func getUserSecurityGroups(cs *egoscale.Client, names []string) ([]egoscale.UserSecurityGroup, error) {
	us := make([]egoscale.UserSecurityGroup, 0, len(names))
	for _, sg := range names {
		s, err := getSecurityGroupByNameOrID(cs, sg)
		if err != nil {
			return nil, err
		}

		us = append(us, s.UserSecurityGroup())
	}
	return us, nil
}

func getDefaultRule(ruleName string, isIpv6 bool) (*egoscale.AuthorizeSecurityGroupIngress, error) {

	icmpType := uint8(8)
	cidr := defaultCidr
	if isIpv6 {
		cidr = defaultCidr6
		icmpType = uint8(128)
	}

	ruleName = strings.ToLower(ruleName)
	if ruleName == "ping" {
		return &egoscale.AuthorizeSecurityGroupIngress{
			Protocol:    "icmp",
			CidrList:    []string{cidr},
			IcmpType:    icmpType,
			IcmpCode:    0,
			Description: "",
		}, nil
	}

	for d := Daytime; d <= Minecraft; d++ {
		if strings.ToLower(port.String(d)) == ruleName {
			return &egoscale.AuthorizeSecurityGroupIngress{
				Protocol:    "tcp",
				CidrList:    []string{cidr},
				StartPort:   uint16(d),
				EndPort:     uint16(d),
				Description: fmt.Sprintf(""),
			}, nil
		}
	}

	return nil, fmt.Errorf("default rule %q not found", ruleName)
}

func addRule(rule *egoscale.AuthorizeSecurityGroupIngress, isEgress bool) error {
	var err error
	if isEgress {
		_, err = cs.Request((*egoscale.AuthorizeSecurityGroupEgress)(rule))
	} else {
		_, err = cs.Request(rule)
	}

	return err
}
