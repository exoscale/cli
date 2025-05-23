package elastic_ip

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
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
	var zones []string
	ctx := gContext

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = utils.AllZones
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
	err := utils.ForEachZone(zones, func(zone string) error {
<<<<<<< Updated upstream:cmd/elastic_ip_list.go
		client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}
=======
		ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))
>>>>>>> Stashed changes:cmd/compute/elastic_ip/elastic_ip_list.go

		list, err := client.ListElasticIPS(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Elastic IP addresses in zone %s: %w", zone, err)
		}

		if list != nil {
			for _, e := range list.ElasticIPS {
				res <- elasticIPListItemOutput{
					ID:        e.ID.String(),
					IPAddress: e.IP,
					Zone:      zone,
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
