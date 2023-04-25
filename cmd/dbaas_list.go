package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasServiceListItemOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Plan string `json:"plan"`
	Zone string `json:"zone"`
}

type dbaasServiceListOutput []dbaasServiceListItemOutput

func (o *dbaasServiceListOutput) toJSON()  { output.JSON(o) }
func (o *dbaasServiceListOutput) toText()  { output.Text(o) }
func (o *dbaasServiceListOutput) toTable() { output.Table(o) }

type dbaasServiceListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *dbaasServiceListCmd) cmdAliases() []string { return gListAlias }

func (c *dbaasServiceListCmd) cmdShort() string { return "List Database Services" }

func (c *dbaasServiceListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Database Services.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&dbaasServiceListItemOutput{}), ", "))
}

func (c *dbaasServiceListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
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
	err := forEachZone(zones, func(zone string) error {
		ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

		list, err := globalstate.GlobalEgoscaleClient.ListDatabaseServices(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Database Services in zone %s: %w", zone, err)
		}

		for _, dbService := range list {
			res <- dbaasServiceListItemOutput{
				Name: *dbService.Name,
				Type: *dbService.Type,
				Plan: *dbService.Plan,
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

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasCmd, &dbaasServiceListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
