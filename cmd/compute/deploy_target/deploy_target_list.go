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
	Zone string `json:"zone"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type deployTargetListOutput []deployTargetListItemOutput

func (o *deployTargetListOutput) ToJSON()  { output.JSON(o) }
func (o *deployTargetListOutput) ToText()  { output.Text(o) }
func (o *deployTargetListOutput) ToTable() { output.Table(o) }

type deployTargetListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
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

	out := make(deployTargetListOutput, 0)
	res := make(chan deployTargetListItemOutput)
	done := make(chan struct{})

	go func() {
		for dt := range res {
			out = append(out, dt)
		}
		done <- struct{}{}
	}()
	err := utils.ForEachZone(zones, func(zone v3.ZoneName) error {

		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}

		list, err := client.ListDeployTargets(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Deploy Targets in zone %s: %w", zone, err)
		}

		for _, dt := range list.DeployTargets {
			res <- deployTargetListItemOutput{
				ID:   dt.ID.String(),
				Name: dt.Name,
				Type: string(dt.Type),
				Zone: string(zone),
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
