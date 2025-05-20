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

type privateNetworkListItemOutput struct {
	ID   v3.UUID     `json:"id"`
	Name string      `json:"name"`
	Zone v3.ZoneName `json:"zone"`
}

type privateNetworkListOutput []privateNetworkListItemOutput

func (o *privateNetworkListOutput) ToJSON()  { output.JSON(o) }
func (o *privateNetworkListOutput) ToText()  { output.Text(o) }
func (o *privateNetworkListOutput) ToTable() { output.Table(o) }

type privateNetworkListCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *privateNetworkListCmd) CmdAliases() []string { return GListAlias }

func (c *privateNetworkListCmd) CmdShort() string { return "List Private Networks" }

func (c *privateNetworkListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Private Networks.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&privateNetworkListItemOutput{}), ", "))
}

func (c *privateNetworkListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := GContext

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

	out := make(privateNetworkListOutput, 0)
	res := make(chan privateNetworkListItemOutput)
	done := make(chan struct{})

	go func() {
		for pn := range res {
			out = append(out, pn)
		}
		done <- struct{}{}
	}()
	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {

		c := client.WithEndpoint((zone.APIEndpoint))
		resp, err := c.ListPrivateNetworks(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Private Networks in zone %s: %w", zone, err)
		}

		for _, p := range resp.PrivateNetworks {
			res <- privateNetworkListItemOutput{
				ID:   p.ID,
				Name: p.Name,
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
	cobra.CheckErr(RegisterCLICommand(privateNetworkCmd, &privateNetworkListCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
