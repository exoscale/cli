package load_balancer

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

type nlbListItemOutput struct {
	ID        v3.UUID     `json:"id"`
	Name      string      `json:"name"`
	Zone      v3.ZoneName `json:"zone"`
	IPAddress string      `json:"ip_address"`
}

type nlbListOutput []nlbListItemOutput

func (o *nlbListOutput) ToJSON()  { output.JSON(o) }
func (o *nlbListOutput) ToText()  { output.Text(o) }
func (o *nlbListOutput) ToTable() { output.Table(o) }

type nlbListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *nlbListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *nlbListCmd) CmdShort() string { return "List Network Load Balancers" }

func (c *nlbListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Network Load Balancers.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbListItemOutput{}), ", "))
}

func (c *nlbListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	var zones []v3.ZoneName
	ctx := exocmd.GContext

	if c.Zone != "" {
		zones = []v3.ZoneName{v3.ZoneName(c.Zone)}
	} else {
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
		if err != nil {
			return err
		}
		zones, err = utils.AllZonesV3(ctx, *client)
		if err != nil {
			return err
		}
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
	err := utils.ForEachZone(zones, func(zone v3.ZoneName) error {
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}

		list, err := client.ListLoadBalancers(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Network Load Balancers in zone %s: %w", zone, err)
		}

		for _, nlb := range list.LoadBalancers {
			res <- nlbListItemOutput{
				ID:        nlb.ID,
				Name:      nlb.Name,
				Zone:      v3.ZoneName(zone),
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

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
