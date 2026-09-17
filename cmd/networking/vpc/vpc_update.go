package vpc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	VPC string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`

	Name         string            `cli-usage:"VPC name"`
	Description  string            `cli-usage:"VPC description"`
	Labels       map[string]string `cli-flag:"label" cli-usage:"VPC label (format: key=value), clearing the labels is possible by passing [=]"`
	Zone         v3.ZoneName       `cli-short:"z" cli-usage:"VPC zone"`
	DNSServers   []string          `cli-flag:"dns-server" cli-usage:"DHCP option 6: DNS servers (can be specified multiple times)"`
	NTPServers   []string          `cli-flag:"ntp-server" cli-usage:"DHCP option 42: NTP servers (can be specified multiple times)"`
	DomainSearch []string          `cli-flag:"domain-search" cli-usage:"DHCP option 119: domain search list (can be specified multiple times)"`
}

func (c *vpcUpdateCmd) CmdAliases() []string { return nil }

func (c *vpcUpdateCmd) CmdShort() string { return "Update a VPC" }

func (c *vpcUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Virtual Private Cloud.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcShowOutput{}), ", "))
}

func (c *vpcUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	entry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	req := v3.UpdateVpcRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		req.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		req.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		req.Labels = exocmd.ConvertIfSpecialEmptyMap(c.Labels)
		updated = true
	}

	// Modify existing DHCP options if any of the DHCP options flags are set
	dhcpOptions := entry.DHCPOptions
	dhcpOptionsChanged := false

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.DNSServers)) {
		dnsServersParsed, err := stringsToIPv4s(c.DNSServers)
		if err != nil {
			return fmt.Errorf("invalid DNS server: %w", err)
		}
		dhcpOptions.DNSServers = dnsServersParsed

		dhcpOptionsChanged = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.NTPServers)) {
		ntpServersParsed, err := stringsToIPv4s(c.NTPServers)
		if err != nil {
			return fmt.Errorf("invalid NTP server: %w", err)
		}
		dhcpOptions.NtpServers = ntpServersParsed

		dhcpOptionsChanged = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.DomainSearch)) {
		dhcpOptions.DomainSearch = c.DomainSearch
		dhcpOptionsChanged = true
	}

	if dhcpOptionsChanged {
		req.DHCPOptions = dhcpOptions
		updated = true
	}

	if updated {
		if _, err := client.UpdateVpc(ctx, entry.ID, req); err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&vpcShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			VPC:                entry.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &vpcUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
