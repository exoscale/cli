package instance_pool

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

type instancePoolListItemOutput struct {
	ID    v3.UUID     `json:"id" outputWidth:"36"`
	Name  string      `json:"name" outputWidth:"70"`
	Zone  v3.ZoneName `json:"zone" outputWidth:"8"`
	Size  int64       `json:"size" outputWidth:"4"`
	State string      `json:"state" outputWidth:"12"`
}

type instancePoolListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *instancePoolListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *instancePoolListCmd) CmdShort() string { return "List Instance Pools" }

func (c *instancePoolListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Instance Pools.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instancePoolListItemOutput{}), ", "))
}

func (c *instancePoolListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runInstancePoolList(c, os.Stdout, os.Stderr)
}

func runInstancePoolList(c *instancePoolListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(instancePoolListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			list, err := zc.ListInstancePools(ctx)
			if err != nil {
				return fmt.Errorf("unable to list Instance Pools in zone %s: %w", zone, err)
			}
			for _, i := range list.InstancePools {
				if err := streamer.Push(instancePoolListItemOutput{
					ID:    i.ID,
					Name:  i.Name,
					Zone:  zone.Name,
					Size:  i.Size,
					State: string(i.State),
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePoolCmd, &instancePoolListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
