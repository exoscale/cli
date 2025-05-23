package deploy_target

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = utils.AllZones
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
	err := utils.ForEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

		list, err := globalstate.EgoscaleClient.ListDeployTargets(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Deploy Targets in zone %s: %w", zone, err)
		}

		for _, dt := range list {
			res <- deployTargetListItemOutput{
				ID:   *dt.ID,
				Name: *dt.Name,
				Type: *dt.Type,
				Zone: zone,
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
