package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
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
	var zones []string

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
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

		list, err := globalstate.EgoscaleClient.ListElasticIPs(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Elastic IP addresses in zone %s: %w", zone, err)
		}

		for _, e := range list {
			res <- elasticIPListItemOutput{
				ID:        *e.ID,
				IPAddress: e.IPAddress.String(),
				Zone:      zone,
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
	cobra.CheckErr(registerCLICommand(elasticIPCmd, &elasticIPListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
