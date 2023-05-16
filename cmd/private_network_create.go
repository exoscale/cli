package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type privateNetworkCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Description string `cli-usage:"Private Network description"`
	EndIP       string `cli-usage:"managed Private Network range end IP address"`
	Netmask     string `cli-usage:"managed Private Network netmask"`
	StartIP     string `cli-usage:"managed Private Network range start IP address"`
	Zone        string `cli-short:"z" cli-usage:"Private Network zone"`
}

func (c *privateNetworkCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *privateNetworkCreateCmd) cmdShort() string {
	return "Create a Private Network"
}

func (c *privateNetworkCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Private Network.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&privateNetworkShowOutput{}), ", "))
}

func (c *privateNetworkCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	privateNetwork := &egoscale.PrivateNetwork{
		Description: utils.NonEmptyStringPtr(c.Description),
		EndIP: func() (v *net.IP) {
			if c.EndIP != "" {
				ip := net.ParseIP(c.EndIP)
				v = &ip
			}
			return
		}(),
		Name: &c.Name,
		Netmask: func() (v *net.IP) {
			if c.Netmask != "" {
				ip := net.ParseIP(c.Netmask)
				v = &ip
			}
			return
		}(),
		StartIP: func() (v *net.IP) {
			if c.StartIP != "" {
				ip := net.ParseIP(c.StartIP)
				v = &ip
			}
			return
		}(),
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Creating Private Network %q...", c.Name), func() {
		privateNetwork, err = globalstate.EgoscaleClient.CreatePrivateNetwork(ctx, c.Zone, privateNetwork)
	})
	if err != nil {
		return err
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
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
