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
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = utils.AllZones
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
	err := utils.ForEachZone(zones, func(zone string) error {
<<<<<<< Updated upstream:cmd/dbaas_list.go
		ctx := gContext
		client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
		if err != nil {
			return err
		}
=======
		ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_list.go

		list, err := client.ListDBAASServices(ctx)
		if err != nil {
			return fmt.Errorf("unable to list Database Services in zone %s: %w", zone, err)
		}

		for _, dbService := range list.DBAASServices {
			res <- dbaasServiceListItemOutput{
				Name: string(dbService.Name),
				Type: string(dbService.Type),
				Plan: dbService.Plan,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceListCmd{
		cliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
