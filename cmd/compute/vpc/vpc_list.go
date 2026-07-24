package vpc

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

type vpcListItemOutput struct {
	ID          v3.UUID     `json:"id" outputWidth:"36"`
	Name        string      `json:"name" outputWidth:"30"`
	Zone        v3.ZoneName `json:"zone" outputWidth:"8"`
	Description string      `json:"description" outputWidth:"40"`
}

type vpcListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *vpcListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *vpcListCmd) CmdShort() string { return "List VPCs" }

func (c *vpcListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Virtual Private Clouds.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcListItemOutput{}), ", "))
}

func (c *vpcListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runVPCList(c, os.Stdout, os.Stderr)
}

func runVPCList(c *vpcListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(vpcListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			resp, err := zc.ListVpcs(ctx)
			if err != nil {
				return fmt.Errorf("unable to list VPCs in zone %s: %w", zone.Name, err)
			}
			for _, v := range resp.Vpcs {
				if err := streamer.Push(vpcListItemOutput{
					ID:          v.ID,
					Name:        v.Name,
					Zone:        zone.Name,
					Description: v.Description,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &vpcListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
