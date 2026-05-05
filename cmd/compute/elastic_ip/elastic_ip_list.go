package elastic_ip

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

type elasticIPListItemOutput struct {
	ID          v3.UUID     `json:"id" outputWidth:"36"`
	IPAddress   string      `json:"ip_address" outputWidth:"18"`
	Description string      `json:"description" outputWidth:"70"`
	Type        string      `json:"type" outputWidth:"10"`
	Zone        v3.ZoneName `json:"zone" outputWidth:"8"`
}

type elasticIPListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *elasticIPListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *elasticIPListCmd) CmdShort() string { return "List Elastic IPs" }

func (c *elasticIPListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Compute Elastic IPs.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPListItemOutput{}), ", "))
}

func (c *elasticIPListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runElasticIPList(c, os.Stdout, os.Stderr)
}

func runElasticIPList(c *elasticIPListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(elasticIPListItemOutput{}, stdout)
	defer func() { _ = streamer.Close() }()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			list, err := zc.ListElasticIPS(ctx)
			if err != nil {
				return fmt.Errorf("unable to list Elastic IP addresses in zone %s: %w", zone, err)
			}
			if list != nil {
				for _, e := range list.ElasticIPS {
					eipType := "Manual"
					if e.Healthcheck != nil {
						eipType = "Managed"
					}
					if err := streamer.Push(elasticIPListItemOutput{
						ID:          e.ID,
						IPAddress:   e.IP,
						Description: e.Description,
						Type:        eipType,
						Zone:        zone.Name,
					}); err != nil {
						return err
					}
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
	cobra.CheckErr(exocmd.RegisterCLICommand(elasticIPCmd, &elasticIPListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
