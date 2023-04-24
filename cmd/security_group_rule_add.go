package cmd

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var securityGroupRuleProtocols = []string{
	"ah",
	"esp",
	"gre",
	"icmp",
	"icmpv6",
	"ipip",
	"tcp",
	"udp",
}

type securityGroupAddRuleCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`

	Description               string `cli-usage:"rule description"`
	FlowDirection             string `cli-flag:"flow" cli-usage:"rule network flow direction (ingress|egress)"`
	ICMPCode                  int64  `cli-usage:"rule ICMP code"`
	ICMPType                  int64  `cli-usage:"rule ICMP type"`
	Port                      string `cli-usage:"rule network port (format: PORT|START-END)"`
	Protocol                  string `cli-usage:"rule network protocol"`
	TargetNetwork             string `cli-flag:"network" cli-usage:"rule target network address (in CIDR format)"`
	TargetSecurityGroup       string `cli-flag:"security-group" cli-usage:"rule target Security Group NAME|ID"`
	TargetPublicSecurityGroup string `cli-flag:"public-security-group" cli-usage:"rule target Public Security Group NAME"`
}

func (c *securityGroupAddRuleCmd) cmdAliases() []string { return nil }

func (c *securityGroupAddRuleCmd) cmdShort() string {
	return "Add a Security Group rule"
}

func (c *securityGroupAddRuleCmd) cmdLong() string {
	return fmt.Sprintf(`This command adds a rule to a Compute instance Security Group.

Supported network protocols: %s

Supported output template annotations: %s`,
		strings.Join(securityGroupRuleProtocols, ", "),
		strings.Join(output.OutputterTemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupAddRuleCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupAddRuleCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	securityGroup, err := cs.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	c.Protocol = strings.ToLower(c.Protocol)
	if !utils.IsInList(securityGroupRuleProtocols, c.Protocol) {
		return fmt.Errorf("unsupported network protocol %q", c.Protocol)
	}

	securityGroupRule := &egoscale.SecurityGroupRule{
		Description:   utils.NonEmptyStringPtr(c.Description),
		FlowDirection: &c.FlowDirection,
		Protocol:      &c.Protocol,
	}

	if (c.TargetNetwork == "" && c.TargetSecurityGroup == "" && c.TargetPublicSecurityGroup == "") ||
		(c.TargetNetwork != "" && c.TargetSecurityGroup != "") ||
		(c.TargetNetwork != "" && c.TargetPublicSecurityGroup != "") ||
		(c.TargetSecurityGroup != "" && c.TargetPublicSecurityGroup != "") {
		return fmt.Errorf("either a target network address or Security Group name/ID must be specified")
	}

	if c.TargetSecurityGroup != "" {
		targetSecurityGroup, err := cs.FindSecurityGroup(ctx, zone, c.TargetSecurityGroup)
		if err != nil {
			return fmt.Errorf("unable to retrieve Security Group %q: %w", c.TargetSecurityGroup, err)
		}
		securityGroupRule.SecurityGroupID = targetSecurityGroup.ID
	} else if c.TargetPublicSecurityGroup != "" {

		visibility := "public"
		securityGroupRule.Visibility = &visibility
		securityGroupRule.SecurityGroupName = &c.TargetPublicSecurityGroup
	} else {
		_, network, err := net.ParseCIDR(c.TargetNetwork)
		if err != nil {
			return fmt.Errorf("invalid value for network %q: %w", c.TargetNetwork, err)
		}
		securityGroupRule.Network = network
	}

	if c.Port != "" {
		startPort, endPort, err := func(portSpec string) (uint16, uint16, error) {
			parts := strings.SplitN(portSpec, "-", 2)
			if len(parts) == 2 {
				s, err := strconv.ParseUint(parts[0], 10, 32)
				if err != nil {
					return 0, 0, err
				}

				e, err := strconv.ParseUint(parts[1], 10, 32)
				if err != nil {
					return 0, 0, err
				}

				return uint16(s), uint16(e), nil
			}

			p, err := strconv.ParseUint(parts[0], 10, 32)
			if err != nil {
				return 0, 0, err
			}

			return uint16(p), uint16(p), nil
		}(c.Port)
		if err != nil {
			return fmt.Errorf("invalid port value %q: %w", c.Port, err)
		}

		for _, v := range []uint16{startPort, endPort} {
			if v < 1 || v > uint16(65535) {
				return errors.New("a port value must be between 1 and 65535")
			}
		}

		if endPort < startPort {
			return fmt.Errorf("end port must be greater than start port")
		}

		securityGroupRule.StartPort = &startPort
		securityGroupRule.EndPort = &endPort
	}

	if strings.HasPrefix(c.Protocol, "icmp") {
		securityGroupRule.ICMPCode = &c.ICMPCode
		securityGroupRule.ICMPType = &c.ICMPType
	}

	decorateAsyncOperation(fmt.Sprintf("Adding rule to Security Group %q...", *securityGroup.Name), func() {
		_, err = cs.CreateSecurityGroupRule(ctx, zone, securityGroup, securityGroupRule)
	})
	if err != nil {
		return err
	}

	return (&securityGroupShowCmd{
		cliCommandSettings: c.cliCommandSettings,
		SecurityGroup:      *securityGroup.ID,
	}).cmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupRuleCmd, &securityGroupAddRuleCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		FlowDirection: "ingress",
		Protocol:      "tcp",
	}))
}
