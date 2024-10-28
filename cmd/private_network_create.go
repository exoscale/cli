package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type privateNetworkCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Description string      `cli-usage:"Private Network description"`
	EndIP       string      `cli-usage:"managed Private Network range end IP address"`
	Netmask     string      `cli-usage:"managed Private Network netmask"`
	StartIP     string      `cli-usage:"managed Private Network range start IP address"`
	Zone        v3.ZoneName `cli-short:"z" cli-usage:"Private Network zone"`
	Option      []string    `cli-usage:"DHCP network option (format: option1=\"value1 value2\")" cli-flag-multi:"true"`
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
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	req := v3.CreatePrivateNetworkRequest{
		Description: func() string {
			if c.Description != "" {
				return *utils.NonEmptyStringPtr(c.Description)
			}
			return ""
		}(),
		EndIP: func() net.IP {
			if c.EndIP != "" {
				return net.ParseIP(c.EndIP)
			}
			return nil
		}(),
		Name: c.Name,
		Netmask: func() net.IP {
			if c.Netmask != "" {
				return net.ParseIP(c.Netmask)
			}
			return nil
		}(),
		StartIP: func() net.IP {
			if c.StartIP != "" {
				return net.ParseIP(c.StartIP)
			}
			return nil
		}(),
	}

	if len(c.Option) > 0 {
		opts, err := processPrivateNetworkOptions(c.Option)
		if err != nil {
			return err
		}
		req.Options = opts
	}

	op, err := client.CreatePrivateNetwork(ctx, req)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating Private Network %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&privateNetworkShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			PrivateNetwork:     c.Name,
			Zone:               v3.ZoneName(c.Zone),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
