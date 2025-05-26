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

type privateNetworkCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Description  string      `cli-usage:"Private Network description"`
	EndIP        string      `cli-usage:"Private Network range end IP address"`
	StartIP      string      `cli-usage:"Private Network range start IP address"`
	Zone         v3.ZoneName `cli-short:"z" cli-usage:"Private Network zone"`
	Netmask      string      `cli-usage:"DHCP option 1: Subnet netmask"`
	DNSServers   []string    `cli-flag:"dns-server" cli-usage:"DHCP option 6: DNS servers (can be specified multiple times)"`
	NTPServers   []string    `cli-flag:"ntp-server" cli-usage:"DHCP option 42: NTP servers (can be specified multiple times)"`
	Routers      []string    `cli-flag:"router" cli-usage:"DHCP option 3: Routers (can be specified multiple times)"`
	DomainSearch []string    `cli-usage:"DHCP option 119: domain search list (limited to 255 octets, can be specified multiple times)"`
}

func (c *privateNetworkCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *privateNetworkCreateCmd) CmdShort() string {
	return "Create a Private Network"
}

func (c *privateNetworkCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Private Network.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&privateNetworkShowOutput{}), ", "))
}

func (c *privateNetworkCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	req := v3.CreatePrivateNetworkRequest{
		Name:        c.Name,
		Description: c.Description,
		EndIP:       net.ParseIP(c.EndIP),
		Netmask:     net.ParseIP(c.Netmask),
		StartIP:     net.ParseIP(c.StartIP),
	}

	opts := &v3.PrivateNetworkOptions{}

	if len(c.DNSServers) > 0 {
		for _, server := range c.DNSServers {
			if ip := net.ParseIP(server); ip != nil {
				opts.DNSServers = append(opts.DNSServers, ip)
			} else {
				return fmt.Errorf("invalid DNS server IP address: %q", server)
			}
		}
	}

	if len(c.NTPServers) > 0 {
		for _, server := range c.NTPServers {
			if ip := net.ParseIP(server); ip != nil {
				opts.NtpServers = append(opts.NtpServers, ip)
			} else {
				return fmt.Errorf("invalid NTP server IP address: %q", server)
			}
		}
	}

	if len(c.Routers) > 0 {
		for _, router := range c.Routers {
			if ip := net.ParseIP(router); ip != nil {
				opts.Routers = append(opts.Routers, ip)
			} else {
				return fmt.Errorf("invalid router IP address: %q", router)
			}
		}
	}

	if len(c.DomainSearch) > 0 {
		opts.DomainSearch = c.DomainSearch
	}

	req.Options = opts

	op, err := client.CreatePrivateNetwork(ctx, req)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Private Network %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&privateNetworkShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			PrivateNetwork:     op.Reference.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(privateNetworkCmd, &privateNetworkCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
