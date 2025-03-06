package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksNodepoolListItemOutput struct {
	ID      v3.UUID     `json:"id"`
	Name    string      `json:"name"`
	Cluster string      `json:"cluster"`
	Size    int64       `json:"size"`
	State   string      `json:"state"`
	Zone    v3.ZoneName `json:"zone"`
}

type sksNodepoolListOutput []sksNodepoolListItemOutput

func (o *sksNodepoolListOutput) ToJSON()  { output.JSON(o) }
func (o *sksNodepoolListOutput) ToText()  { output.Text(o) }
func (o *sksNodepoolListOutput) ToTable() { output.Table(o) }

type sksNodepoolListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *sksNodepoolListCmd) cmdAliases() []string { return gListAlias }

func (c *sksNodepoolListCmd) cmdShort() string { return "List SKS cluster Nodepools" }

func (c *sksNodepoolListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists SKS cluster Nodepools.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksNodepoolListItemOutput{}), ", "))
}

func (c *sksNodepoolListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := gContext

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

	out := make(sksNodepoolListOutput, 0)
	res := make(chan sksNodepoolListItemOutput)
	done := make(chan struct{})

	go func() {
		for cluster := range res {
			out = append(out, cluster)
		}
		done <- struct{}{}
	}()
	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {

		c := client.WithEndpoint((zone.APIEndpoint))
		listResp, err := c.ListSKSClusters(ctx)
		if err != nil {
			return fmt.Errorf("unable to list SKS clusters in zone %s: %w", zone, err)
		}

		for _, cluster := range listResp.SKSClusters {
			for _, np := range cluster.Nodepools {
				res <- sksNodepoolListItemOutput{
					ID:      np.ID,
					Name:    np.Name,
					Cluster: cluster.Name,
					Size:    np.Size,
					State:   string(np.State),
					Zone:    zone.Name,
				}
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

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
