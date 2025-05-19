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
	ctx := gContext

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
		client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}

		list, err := client.ListLoadBalancers(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Network Load Balancers in zone %s: %w", zone, err)
		}

		for _, nlb := range list.LoadBalancers {
			res <- nlbListItemOutput{
				ID:        nlb.ID.String(),
				Name:      nlb.Name,
				Zone:      zone,
				IPAddress: nlb.IP.String(),
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
