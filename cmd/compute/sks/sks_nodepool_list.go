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

type sksNodepoolListItemOutput struct {
	ID      v3.UUID     `json:"id" outputWidth:"36"`
	Name    string      `json:"name" outputWidth:"70"`
	Cluster string      `json:"cluster" outputWidth:"38"`
	Size    int64       `json:"size" outputWidth:"4"`
	State   string      `json:"state" outputWidth:"12"`
	Zone    v3.ZoneName `json:"zone" outputWidth:"8"`
}

type sksNodepoolListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *sksNodepoolListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *sksNodepoolListCmd) CmdShort() string { return "List SKS cluster Nodepools" }

func (c *sksNodepoolListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists SKS cluster Nodepools.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolListItemOutput{}), ", "))
}

func (c *sksNodepoolListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runSksNodepoolList(c, os.Stdout, os.Stderr)
}

func runSksNodepoolList(c *sksNodepoolListCmd, stdout, stderr io.Writer) error {
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

	streamer := output.NewStreamer(sksNodepoolListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			listResp, err := zc.ListSKSClusters(ctx)
			if err != nil {
				return fmt.Errorf("unable to list SKS clusters in zone %s: %w", zone, err)
			}
			for _, cluster := range listResp.SKSClusters {
				for _, np := range cluster.Nodepools {
					if err := streamer.Push(sksNodepoolListItemOutput{
						ID:      np.ID,
						Name:    np.Name,
						Cluster: cluster.Name,
						Size:    np.Size,
						State:   string(np.State),
						Zone:    zone.Name,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(sksNodepoolCmd, &sksNodepoolListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
