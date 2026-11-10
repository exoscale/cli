package private_network

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type privateNetworkUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description  string      `cli-usage:"Private Network description"`
	EndIP        string      `cli-usage:"Private Network range end IP address"`
	Name         string      `cli-usage:"Private Network name"`
	StartIP      string      `cli-usage:"Private Network range start IP address"`
	Zone         v3.ZoneName `cli-short:"z" cli-usage:"Private Network zone"`
	Netmask      string      `cli-usage:"DHCP option 1: Subnet netmask"`
	DNSServers   []string    `cli-flag:"dns-server" cli-usage:"DHCP option 6: DNS servers (can be specified multiple times)"`
	NTPServers   []string    `cli-flag:"ntp-server" cli-usage:"DHCP option 42: NTP servers (can be specified multiple times)"`
	Routers      []string    `cli-flag:"router" cli-usage:"DHCP option 3: Routers (can be specified multiple times)"`
	DomainSearch []string    `cli-usage:"DHCP option 119: domain search list (limited to 255 octets, can be specified multiple times)"`
}

func (c *privateNetworkUpdateCmd) CmdAliases() []string { return nil }

func (c *privateNetworkUpdateCmd) CmdShort() string { return "Update a Private Network" }

func (c *privateNetworkUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Compute instance Private Network.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&privateNetworkShowOutput{}), ", "),
	)
}

func (c *privateNetworkUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListPrivateNetworks(ctx)
	if err != nil {
		return err
	}

	pn, err := resp.FindPrivateNetwork(c.PrivateNetwork)
	if err != nil {
		return err
	}

	updateReq := v3.UpdatePrivateNetworkRequest{}
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		updateReq.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.EndIP)) {
		ip := net.ParseIP(c.EndIP)
		updateReq.EndIP = ip
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		updateReq.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Netmask)) {
		ip := net.ParseIP(c.Netmask)
		updateReq.Netmask = ip
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.StartIP)) {
		ip := net.ParseIP(c.StartIP)
		updateReq.StartIP = ip
		updated = true
	}

	opts := pn.Options
	if opts == nil {
		opts = &v3.PrivateNetworkOptions{}
	}

	optionsChanged := false

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.DNSServers)) {
		opts.DNSServers = nil // Reset before adding new values
		for _, server := range c.DNSServers {
			if ip := net.ParseIP(server); ip != nil {
				opts.DNSServers = append(opts.DNSServers, ip)
			} else {
				return fmt.Errorf("invalid DNS server IP address: %q", server)
			}
		}
		optionsChanged = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.NTPServers)) {
		opts.NtpServers = nil // Reset before adding new values
		for _, server := range c.NTPServers {
			if ip := net.ParseIP(server); ip != nil {
				opts.NtpServers = append(opts.NtpServers, ip)
			} else {
				return fmt.Errorf("invalid NTP server IP address: %q", server)
			}
		}
		optionsChanged = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Routers)) {
		opts.Routers = nil // Reset before adding new values
		for _, router := range c.Routers {
			if ip := net.ParseIP(router); ip != nil {
				opts.Routers = append(opts.Routers, ip)
			} else {
				return fmt.Errorf("invalid router IP address: %q", router)
			}
		}
		optionsChanged = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.DomainSearch)) {
		opts.DomainSearch = c.DomainSearch
		optionsChanged = true
	}

	if optionsChanged {
		updateReq.Options = opts
		updated = true
	}

	var privnetID v3.UUID

	if updated {
		op, err := client.UpdatePrivateNetwork(ctx, pn.ID, updateReq)
		if err != nil {
			return err
		}
		privnetID = op.Reference.ID

		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Private Network %q...", c.PrivateNetwork), func() {
			op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&privateNetworkShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			PrivateNetwork:     privnetID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(privateNetworkCmd, &privateNetworkUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
