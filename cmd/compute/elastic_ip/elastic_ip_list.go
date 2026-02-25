package elastic_ip

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

type elasticIPListItemOutput struct {
	ID        v3.UUID     `json:"id"`
	IPAddress string      `json:"ip_address"`
	Type      string      `json:"type"`
	Zone      v3.ZoneName `json:"zone"`
}

type elasticIPListOutput []elasticIPListItemOutput

func (o *elasticIPListOutput) ToJSON()  { output.JSON(o) }
func (o *elasticIPListOutput) ToText()  { output.Text(o) }
func (o *elasticIPListOutput) ToTable() { output.Table(o) }

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
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	out := make(elasticIPListOutput, 0)
	res := make(chan elasticIPListItemOutput)
	done := make(chan struct{})

	go func() {
		for nlb := range res {
			out = append(out, nlb)
		}
		done <- struct{}{}
	}()
	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {
		c := client.WithEndpoint(zone.APIEndpoint)
		list, err := c.ListElasticIPS(ctx)

		if err != nil {
			return fmt.Errorf("unable to list Elastic IP addresses in zone %s: %w", zone, err)
		}

		if list != nil {
			for _, e := range list.ElasticIPS {
				var eipType string
				if e.Healthcheck != nil {
					eipType = "Managed"
				} else {
					eipType = "Manual"
				}
				res <- elasticIPListItemOutput{
					ID:        e.ID,
					IPAddress: e.IP,
					Type:      eipType,
					Zone:      zone.Name,
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

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(elasticIPCmd, &elasticIPListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
