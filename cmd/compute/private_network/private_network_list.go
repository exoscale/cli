package private_network

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type privateNetworkListItemOutput struct {
	ID   v3.UUID     `json:"id" outputWidth:"36"`
	Name string      `json:"name" outputWidth:"38"`
	Zone v3.ZoneName `json:"zone" outputWidth:"8"`
}

type privateNetworkListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *privateNetworkListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *privateNetworkListCmd) CmdShort() string { return "List Private Networks" }

func (c *privateNetworkListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Private Networks.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&privateNetworkListItemOutput{}), ", "))
}

func (c *privateNetworkListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runPrivateNetworkList(c, os.Stdout, os.Stderr)
}

func runPrivateNetworkList(c *privateNetworkListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(privateNetworkListItemOutput{}, stdout)
	defer func() { _ = streamer.Close() }()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			resp, err := zc.ListPrivateNetworks(ctx)
			if err != nil {
				return fmt.Errorf("unable to list Private Networks in zone %s: %w", zone, err)
			}
			for _, p := range resp.PrivateNetworks {
				if err := streamer.Push(privateNetworkListItemOutput{
					ID:   p.ID,
					Name: p.Name,
					Zone: zone.Name,
				}); err != nil {
					return err
				}
			}
			return nil
		})

	if failed > 0 {
		return fmt.Errorf("%d zone(s) failed", failed)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(privateNetworkCmd, &privateNetworkListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
