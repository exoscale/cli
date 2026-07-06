package sks

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

type sksClusterListItemOutput struct {
	ID           v3.UUID     `json:"id" outputWidth:"36"`
	Name         string      `json:"name" outputWidth:"70"`
	Zone         v3.ZoneName `json:"zone" outputWidth:"8"`
	AuditEnabled bool        `json:"audit_enabled" outputWidth:"11"`
}

type sksListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *sksListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *sksListCmd) CmdShort() string { return "List SKS clusters" }

func (c *sksListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists SKS clusters.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksClusterListItemOutput{}), ", "))
}

func (c *sksListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runSksList(c, os.Stdout, os.Stderr)
}

func runSksList(c *sksListCmd, stdout, stderr io.Writer) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	resp, err := client.ListZones(ctx)
	if err != nil {
		return err
	}
	zones := resp.Zones

	if c.Zone != "" {
		endpoint, err := client.GetZoneAPIEndpoint(ctx, c.Zone)
		if err != nil {
			return err
		}
		zones = []v3.Zone{{APIEndpoint: endpoint}}
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(sksClusterListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			resp, err := zc.ListSKSClusters(ctx)
			if err != nil {
				return fmt.Errorf("unable to list SKS clusters in zone %s: %w", zone, err)
			}
			for _, cluster := range resp.SKSClusters {
				if err := streamer.Push(sksClusterListItemOutput{
					ID:   cluster.ID,
					Name: cluster.Name,
					Zone: zone.Name,
					AuditEnabled: func() bool {
						return cluster.Audit != nil && cluster.Audit.Enabled != nil && *cluster.Audit.Enabled
					}(),
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
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
