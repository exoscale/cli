package dbaas

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

type dbaasServiceListItemOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Plan string `json:"plan"`
	Zone string `json:"zone"`
}

type dbaasServiceListOutput []dbaasServiceListItemOutput

func (o *dbaasServiceListOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasServiceListOutput) ToText()  { output.Text(o) }
func (o *dbaasServiceListOutput) ToTable() { output.Table(o) }

type dbaasServiceListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *dbaasServiceListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *dbaasServiceListCmd) CmdShort() string { return "List Database Services" }

func (c *dbaasServiceListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Database Services.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&dbaasServiceListItemOutput{}), ", "))
}

func (c *dbaasServiceListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceListCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	out := make(dbaasServiceListOutput, 0)
	res := make(chan dbaasServiceListItemOutput)
	done := make(chan struct{})

	go func() {
		for dbService := range res {
			out = append(out, dbService)
		}
		done <- struct{}{}
	}()
	err := utils.ForEachZone(zones, func(zone v3.ZoneName) error {
		ctx := exocmd.GContext
		client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}

		list, err := client.ListDBAASServices(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Database Services in zone %s: %w", zone, err)
		}

		for _, dbService := range list.DBAASServices {
			res <- dbaasServiceListItemOutput{
				Name: string(dbService.Name),
				Type: string(dbService.Type),
				Plan: dbService.Plan,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
