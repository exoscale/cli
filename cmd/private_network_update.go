package cmd

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type privateNetworkUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description string `cli-usage:"Private Network description"`
	EndIP       string `cli-usage:"managed Private Network range end IP address"`
	Name        string `cli-usage:"Private Network name"`
	Netmask     string `cli-usage:"managed Private Network netmask"`
	StartIP     string `cli-usage:"managed Private Network range start IP address"`
	Zone        string `cli-short:"z" cli-usage:"Private Network zone"`
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

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	privateNetwork, err := globalstate.EgoscaleClient.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		privateNetwork.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.EndIP)) {
		ip := net.ParseIP(c.EndIP)
		privateNetwork.EndIP = &ip
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		privateNetwork.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Netmask)) {
		ip := net.ParseIP(c.Netmask)
		privateNetwork.Netmask = &ip
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.StartIP)) {
		ip := net.ParseIP(c.StartIP)
		privateNetwork.StartIP = &ip
		updated = true
	}

	if updated {
		decorateAsyncOperation(fmt.Sprintf("Updating Private Network %q...", c.PrivateNetwork), func() {
			if err = globalstate.EgoscaleClient.UpdatePrivateNetwork(ctx, c.Zone, privateNetwork); err != nil {
				return
			}
		})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&privateNetworkShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			PrivateNetwork:     *privateNetwork.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
