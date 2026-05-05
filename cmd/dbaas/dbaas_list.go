package dbaas

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

type dbaasServiceListItemOutput struct {
	Name string      `json:"name" outputWidth:"38"`
	Type string      `json:"type" outputWidth:"16"`
	Plan string      `json:"plan" outputWidth:"16"`
	Zone v3.ZoneName `json:"zone" outputWidth:"8"`
}

type dbaasServiceListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *dbaasServiceListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *dbaasServiceListCmd) CmdShort() string { return "List Database Services" }

func (c *dbaasServiceListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Database Services.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&dbaasServiceListItemOutput{}), ", "))
}

func (c *dbaasServiceListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runDbaasList(c, os.Stdout, os.Stderr)
}

func runDbaasList(c *dbaasServiceListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(dbaasServiceListItemOutput{}, stdout)
	defer func() { _ = streamer.Close() }()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			list, err := zc.ListDBAASServices(ctx)
			if err != nil {
				return fmt.Errorf("unable to list Database Services in zone %s: %w", zone, err)
			}
			for _, dbService := range list.DBAASServices {
				if err := streamer.Push(dbaasServiceListItemOutput{
					Name: string(dbService.Name),
					Type: string(dbService.Type),
					Plan: dbService.Plan,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
