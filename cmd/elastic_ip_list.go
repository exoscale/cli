package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type elasticIPListItemOutput struct {
	ID          v3.UUID     `json:"id"`
	IPAddress   string      `json:"ip_address"`
	Zone        v3.ZoneName `json:"zone"`
	Description string      `json:description`
}

type elasticIPListOutput []elasticIPListItemOutput

func (o *elasticIPListOutput) ToJSON()  { output.JSON(o) }
func (o *elasticIPListOutput) ToText()  { output.Text(o) }
func (o *elasticIPListOutput) ToTable() { output.Table(o) }

type elasticIPListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *elasticIPListCmd) cmdAliases() []string { return gListAlias }

func (c *elasticIPListCmd) cmdShort() string { return "List Elastic IPs" }

func (c *elasticIPListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Compute Elastic IPs.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&elasticIPListItemOutput{}), ", "))
}

func (c *elasticIPListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *elasticIPListCmd) cmdRun(_ *cobra.Command, _ []string) error {
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

	output := make(elasticIPListOutput, 0)

	for _, zone := range zones {
		c := client.WithEndpoint(zone.APIEndpoint)

		resp, err := c.ListElasticIPS(ctx)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr,
				"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
			continue
		}

		for _, elasticIP := range resp.ElasticIPS {
			output = append(output, elasticIPListItemOutput{
				ID:          elasticIP.ID,
				IPAddress:   elasticIP.IP,
				Zone:        zone.Name,
				Description: elasticIP.Description,
			})
		}
	}

	return c.outputFunc(&output, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
