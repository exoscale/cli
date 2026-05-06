package load_balancer

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

type nlbListItemOutput struct {
	ID        v3.UUID     `json:"id" outputWidth:"36"`
	Name      string      `json:"name" outputWidth:"70"`
	Zone      v3.ZoneName `json:"zone" outputWidth:"8"`
	IPAddress string      `json:"ip_address" outputWidth:"18"`
}

type nlbListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *nlbListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *nlbListCmd) CmdShort() string { return "List Network Load Balancers" }

func (c *nlbListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Network Load Balancers.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbListItemOutput{}), ", "))
}

func (c *nlbListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runNlbList(c, os.Stdout, os.Stderr)
}

func runNlbList(c *nlbListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(nlbListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			list, err := zc.ListLoadBalancers(ctx)
			if err != nil {
				return fmt.Errorf("unable to list Network Load Balancers in zone %s: %w", zone, err)
			}
			for _, nlb := range list.LoadBalancers {
				if err := streamer.Push(nlbListItemOutput{
					ID:        nlb.ID,
					Name:      nlb.Name,
					Zone:      zone.Name,
					IPAddress: nlb.IP.String(),
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
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
