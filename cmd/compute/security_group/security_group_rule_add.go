package security_group

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
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
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`

	Description               string                                        `cli-usage:"rule description"`
	FlowDirection             v3.AddRuleToSecurityGroupRequestFlowDirection `cli-flag:"flow" cli-usage:"rule network flow direction (ingress|egress)"`
	ICMPCode                  int64                                         `cli-usage:"rule ICMP code"`
	ICMPType                  int64                                         `cli-usage:"rule ICMP type"`
	Port                      string                                        `cli-usage:"rule network port (format: PORT|START-END)"`
	Protocol                  v3.AddRuleToSecurityGroupRequestProtocol      `cli-usage:"rule network protocol"`
	TargetNetwork             string                                        `cli-flag:"network" cli-usage:"rule target network address (in CIDR format)"`
	TargetSecurityGroup       string                                        `cli-flag:"security-group" cli-usage:"rule target Security Group NAME|ID"`
	TargetPublicSecurityGroup string                                        `cli-flag:"public-security-group" cli-usage:"rule target Public Security Group NAME"`
}

func (c *securityGroupAddRuleCmd) CmdAliases() []string { return nil }

func (c *securityGroupAddRuleCmd) CmdShort() string {
	return "Add a Security Group rule"
}

func (c *securityGroupAddRuleCmd) CmdLong() string {
	return fmt.Sprintf(`This command adds a rule to a Compute instance Security Group.

Supported network protocols: %s

Supported output template annotations: %s`,
		strings.Join(securityGroupRuleProtocols, ", "),
		strings.Join(output.TemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupAddRuleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupAddRuleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	securityGroups, err := client.ListSecurityGroups(ctx)
	if err != nil {
		return err
	}
	securityGroup, err := securityGroups.FindSecurityGroup(c.SecurityGroup)
	if err != nil {
		return err
	}

	lowercaseProtocol := v3.AddRuleToSecurityGroupRequestProtocol(strings.ToLower(string(c.Protocol)))
	if !utils.IsInList(securityGroupRuleProtocols, string(lowercaseProtocol)) {
		return fmt.Errorf("unsupported network protocol %q", c.Protocol)
	}

	securityGroupRule := &v3.AddRuleToSecurityGroupRequest{
		Description:   c.Description,
		FlowDirection: c.FlowDirection,
		Protocol:      lowercaseProtocol,
	}

	if (c.TargetNetwork == "" && c.TargetSecurityGroup == "" && c.TargetPublicSecurityGroup == "") ||
		(c.TargetNetwork != "" && c.TargetSecurityGroup != "") ||
		(c.TargetNetwork != "" && c.TargetPublicSecurityGroup != "") ||
		(c.TargetSecurityGroup != "" && c.TargetPublicSecurityGroup != "") {
		fmt.Println((c.TargetNetwork == "" && c.TargetSecurityGroup == "" && c.TargetPublicSecurityGroup == ""))
		fmt.Println((c.TargetNetwork != "" && c.TargetSecurityGroup != ""))
		fmt.Println((c.TargetNetwork != "" && c.TargetPublicSecurityGroup != ""))
		fmt.Println((c.TargetSecurityGroup != "" && c.TargetPublicSecurityGroup != ""))
		return fmt.Errorf("either a target network address or Security Group name/ID must be specified")
	}

	if (lowercaseProtocol == "tcp" || lowercaseProtocol == "udp") && c.Port == "" {
		return fmt.Errorf("a port must be specifed for tcp or udp protocol")
	}

	if c.TargetSecurityGroup != "" { //nolint:gocritic
		targetSecurityGroup, err := securityGroups.FindSecurityGroup(c.TargetSecurityGroup)
		if err != nil {
			return fmt.Errorf("unable to retrieve Security Group %q: %w", c.TargetSecurityGroup, err)
		}
		securityGroupRule.SecurityGroup = &v3.SecurityGroupResource{
			ID: targetSecurityGroup.ID,
		}
	} else if c.TargetPublicSecurityGroup != "" {

		securityGroupRule.SecurityGroup = &v3.SecurityGroupResource{
			Visibility: v3.SecurityGroupResourceVisibilityPublic,
			Name:       c.TargetPublicSecurityGroup,
		}

	} else {
		_, network, err := net.ParseCIDR(c.TargetNetwork)
		if err != nil {
			return fmt.Errorf("invalid value for network %q: %w", c.TargetNetwork, err)
		}
		securityGroupRule.Network = network.String()
	}

	if c.Port != "" {
		startPort, endPort, err := func(portSpec string) (int64, int64, error) {
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

				return int64(s), int64(e), nil
			}

			p, err := strconv.ParseUint(parts[0], 10, 32)
			if err != nil {
				return 0, 0, err
			}

			return int64(p), int64(p), nil
		}(c.Port)
		if err != nil {
			return fmt.Errorf("invalid port value %q: %w", c.Port, err)
		}

		for _, v := range []int64{startPort, endPort} {
			if v < 1 || v > int64(65535) {
				return errors.New("a port value must be between 1 and 65535")
			}
		}

		if endPort < startPort {
			return fmt.Errorf("end port must be greater than start port")
		}

		securityGroupRule.StartPort = startPort
		securityGroupRule.EndPort = endPort
	}

	if strings.HasPrefix(string(lowercaseProtocol), "icmp") {
		securityGroupRule.ICMP = &v3.AddRuleToSecurityGroupRequestICMP{
			Code: &c.ICMPCode,
			Type: &c.ICMPType,
		}
	}

	op, err := client.AddRuleToSecurityGroup(ctx, securityGroup.ID, *securityGroupRule)
	if err != nil {
		return err
	}
	exocmd.DecorateAsyncOperation(fmt.Sprintf("Adding rule to Security Group %q...", securityGroup.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&securityGroupShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			SecurityGroup:      securityGroup.ID.String(),
		}).CmdRun(nil, nil)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupRuleCmd, &securityGroupAddRuleCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		FlowDirection: "ingress",
		Protocol:      "tcp",
	}))
}
