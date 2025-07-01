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
	ID        string `json:"id"`
	IPAddress string `json:"ip_address"`
	Zone      string `json:"zone"`
}

type elasticIPListOutput []elasticIPListItemOutput

func (o *elasticIPListOutput) ToJSON()  { output.JSON(o) }
func (o *elasticIPListOutput) ToText()  { output.Text(o) }
func (o *elasticIPListOutput) ToTable() { output.Table(o) }

type elasticIPListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
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
	var zones []v3.ZoneName
	ctx := exocmd.GContext

	if c.Zone != "" {
		zones = []v3.ZoneName{v3.ZoneName(c.Zone)}
	} else {
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
		if err != nil {
			return err
		}
		zones, err = utils.AllZonesV3(ctx, client)
		if err != nil {
			return err
		}
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
	err := utils.ForEachZone(zones, func(zone v3.ZoneName) error {
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}

		list, err := client.ListElasticIPS(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Elastic IP addresses in zone %s: %w", zone, err)
		}

		if list != nil {
			for _, e := range list.ElasticIPS {
				res <- elasticIPListItemOutput{
					ID:        e.ID.String(),
					IPAddress: e.IP,
					Zone:      string(zone),
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
