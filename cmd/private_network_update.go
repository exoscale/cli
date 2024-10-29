package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type privateNetworkUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description  string      `cli-usage:"Private Network description"`
	EndIP        string      `cli-usage:"Private Network range end IP address"`
	Name         string      `cli-usage:"Private Network name"`
	StartIP      string      `cli-usage:"Private Network range start IP address"`
	Zone         v3.ZoneName `cli-short:"z" cli-usage:"Private Network zone"`
	Netmask      string      `cli-usage:"DHCP option 1: Subnet netmask"`
	DNSServers   []string    `cli-usage:"DHCP option 6: DNS servers"`
	NTPServers   []string    `cli-usage:"DHCP option 42: NTP servers"`
	Routers      []string    `cli-usage:"DHCP option 3: Routers"`
	DomainSearch []string    `cli-usage:"DHCP option 119: domain search list (limited to 255 octets)"`
}

func (c *privateNetworkUpdateCmd) cmdAliases() []string { return nil }

func (c *privateNetworkUpdateCmd) cmdShort() string { return "Update a Private Network" }

func (c *privateNetworkUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Compute instance Private Network.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&privateNetworkShowOutput{}), ", "),
	)
}

func (c *privateNetworkUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		updateReq.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.EndIP)) {
		ip := net.ParseIP(c.EndIP)
		updateReq.EndIP = ip
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		updateReq.Name = c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Netmask)) {
		ip := net.ParseIP(c.Netmask)
		updateReq.Netmask = ip
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.StartIP)) {
		ip := net.ParseIP(c.StartIP)
		updateReq.StartIP = ip
		updated = true
	}

	// Process DHCP options if any are changed
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DNSServers)) ||
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.NTPServers)) ||
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Routers)) ||
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.DomainSearch)) {

		opts := &v3.PrivateNetworkOptions{}

		for _, server := range c.DNSServers {
			if ip := net.ParseIP(server); ip != nil {
				opts.DNSServers = append(opts.DNSServers, ip)
			} else {
				return fmt.Errorf("invalid DNS server IP address: %q", server)
			}
		}

		for _, server := range c.NTPServers {
			if ip := net.ParseIP(server); ip != nil {
				opts.NtpServers = append(opts.NtpServers, ip)
			} else {
				return fmt.Errorf("invalid NTP server IP address: %q", server)
			}
		}

		for _, router := range c.Routers {
			if ip := net.ParseIP(router); ip != nil {
				opts.Routers = append(opts.Routers, ip)
			} else {
				return fmt.Errorf("invalid router IP address: %q", router)
			}
		}

		opts.DomainSearch = c.DomainSearch
		updateReq.Options = opts
		updated = true
	}

	if updated {
		op, err := client.UpdatePrivateNetwork(ctx, pn.ID, updateReq)
		if err != nil {
			return err
		}
		decorateAsyncOperation(fmt.Sprintf("Updating Private Network %q...", c.Name), func() {
			op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&privateNetworkShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			PrivateNetwork:     pn.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
