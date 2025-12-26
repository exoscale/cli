package deploy_target

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

type deployTargetListItemOutput struct {
	Zone v3.ZoneName `json:"zone"`
	ID   v3.UUID     `json:"id"`
	Name string      `json:"name"`
	Type string      `json:"type"`
}

type deployTargetListOutput []deployTargetListItemOutput

func (o *deployTargetListOutput) ToJSON()  { output.JSON(o) }
func (o *deployTargetListOutput) ToText()  { output.Text(o) }
func (o *deployTargetListOutput) ToTable() { output.Table(o) }

type deployTargetListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *deployTargetListCmd) CmdAliases() []string { return nil }

func (c *deployTargetListCmd) CmdShort() string { return "List Deploy Targets" }

func (c *deployTargetListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists existing Deploy Targets.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&deployTargetListOutput{}), ", "))
}

func (c *deployTargetListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *deployTargetListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

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

	out := make(deployTargetListOutput, 0)
	res := make(chan deployTargetListItemOutput)
	done := make(chan struct{})

	go func() {
		for dt := range res {
			out = append(out, dt)
		}
		done <- struct{}{}
	}()

	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {
		c := client.WithEndpoint(zone.APIEndpoint)
		list, err := c.ListDeployTargets(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Deploy Targets in zone %s: %w", zone, err)
		}

		for _, dt := range list.DeployTargets {
			res <- deployTargetListItemOutput{
				ID:   dt.ID,
				Name: dt.Name,
				Type: string(dt.Type),
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
	cobra.CheckErr(exocmd.RegisterCLICommand(deployTargetCmd, &deployTargetListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
