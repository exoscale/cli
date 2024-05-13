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

type nlbListItemOutput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Zone      string `json:"zone"`
	IPAddress string `json:"ip_address"`
}

type nlbListOutput []nlbListItemOutput

func (o *nlbListOutput) ToJSON()  { output.JSON(o) }
func (o *nlbListOutput) ToText()  { output.Text(o) }
func (o *nlbListOutput) ToTable() { output.Table(o) }

type nlbListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *nlbListCmd) cmdAliases() []string { return gListAlias }

func (c *nlbListCmd) cmdShort() string { return "List Network Load Balancers" }

func (c *nlbListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Network Load Balancers.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbListItemOutput{}), ", "))
}

func (c *nlbListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = utils.AllZones
	}

	out := make(nlbListOutput, 0)
	res := make(chan nlbListItemOutput)
	done := make(chan struct{})

	go func() {
		for nlb := range res {
			out = append(out, nlb)
		}
		done <- struct{}{}
	}()
	err := utils.ForEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

		list, err := globalstate.EgoscaleClient.ListNetworkLoadBalancers(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Network Load Balancers in zone %s: %w", zone, err)
		}

		for _, nlb := range list {
			res <- nlbListItemOutput{
				ID:        *nlb.ID,
				Name:      *nlb.Name,
				Zone:      zone,
				IPAddress: utils.DefaultIP(nlb.IPAddress, ""),
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
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
