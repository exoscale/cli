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

	Description string      `cli-usage:"Private Network description"`
	EndIP       string      `cli-usage:"managed Private Network range end IP address"`
	Name        string      `cli-usage:"Private Network name"`
	Netmask     string      `cli-usage:"managed Private Network netmask"`
	StartIP     string      `cli-usage:"managed Private Network range start IP address"`
	Zone        v3.ZoneName `cli-short:"z" cli-usage:"Private Network zone"`
	Option      []string    `cli-usage:"DHCP network option (format: option1=\"value1 value2\")" cli-flag-multi:"true"`
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

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Option)) {
		opts := &v3.PrivateNetworkOptions{}
		optionsMap := make(map[string][]string)

		// Process each option flag
		for _, opt := range c.Option {
			keyValue := strings.SplitN(opt, "=", 2)
			if len(keyValue) != 2 {
				continue
			}
			key := keyValue[0]
			values := strings.Split(keyValue[1], " ")
			optionsMap[key] = append(optionsMap[key], values...)
		}

		// Process collected values
		for key, values := range optionsMap {
			switch key {
			case "dns-servers":
				for _, v := range values {
					if ip := net.ParseIP(v); ip != nil {
						opts.DNSServers = append(opts.DNSServers, ip)
					}
				}
			case "ntp-servers":
				for _, v := range values {
					if ip := net.ParseIP(v); ip != nil {
						opts.NtpServers = append(opts.NtpServers, ip)
					}
				}
			case "routers":
				for _, v := range values {
					if ip := net.ParseIP(v); ip != nil {
						opts.Routers = append(opts.Routers, ip)
					}
				}
			case "domain-search":
				opts.DomainSearch = values
			}
		}
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
