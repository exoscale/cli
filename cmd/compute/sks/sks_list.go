package sks

import (
	"fmt"
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
	ID   v3.UUID     `json:"id"`
	Name string      `json:"name"`
	Zone v3.ZoneName `json:"zone"`
}

type sksClusterListOutput []sksClusterListItemOutput

func (o *sksClusterListOutput) ToJSON()  { output.JSON(o) }
func (o *sksClusterListOutput) ToText()  { output.Text(o) }
func (o *sksClusterListOutput) ToTable() { output.Table(o) }

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

	out := make(sksClusterListOutput, 0)
	res := make(chan sksClusterListItemOutput)
	done := make(chan struct{})

	go func() {
		for cluster := range res {
			out = append(out, cluster)
		}
		done <- struct{}{}
	}()
	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {
		c := client.WithEndpoint((zone.APIEndpoint))
		resp, err := c.ListSKSClusters(ctx)
		if err != nil {
			return fmt.Errorf("unable to list SKS clusters in zone %s: %w", zone, err)
		}

		for _, cluster := range resp.SKSClusters {
			res <- sksClusterListItemOutput{
				ID:   cluster.ID,
				Name: cluster.Name,
				Zone: zone.Name,
			}
		}

		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	close(res)
	<-done

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
