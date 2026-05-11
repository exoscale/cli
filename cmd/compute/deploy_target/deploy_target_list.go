package deploy_target

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

type deployTargetListItemOutput struct {
	Zone v3.ZoneName `json:"zone" outputWidth:"8"`
	ID   v3.UUID     `json:"id" outputWidth:"36"`
	Name string      `json:"name" outputWidth:"70"`
	Type string      `json:"type" outputWidth:"16"`
}

type deployTargetListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *deployTargetListCmd) CmdAliases() []string { return nil }

func (c *deployTargetListCmd) CmdShort() string { return "List Deploy Targets" }

func (c *deployTargetListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists existing Deploy Targets.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&deployTargetListItemOutput{}), ", "))
}

func (c *deployTargetListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *deployTargetListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runDeployTargetList(c, os.Stdout, os.Stderr)
}

func runDeployTargetList(c *deployTargetListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(deployTargetListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			list, err := zc.ListDeployTargets(ctx)
			if err != nil {
				return fmt.Errorf("unable to list Deploy Targets in zone %s: %w", zone, err)
			}
			for _, dt := range list.DeployTargets {
				if err := streamer.Push(deployTargetListItemOutput{
					ID:   dt.ID,
					Name: dt.Name,
					Type: string(dt.Type),
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
	cobra.CheckErr(exocmd.RegisterCLICommand(deployTargetCmd, &deployTargetListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
