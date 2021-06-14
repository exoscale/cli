package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbServiceListItemOutput struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Plan string `json:"plan"`
	Zone string `json:"zone"`
}

type dbServiceListOutput []dbServiceListItemOutput

func (o *dbServiceListOutput) toJSON()  { outputJSON(o) }
func (o *dbServiceListOutput) toText()  { outputText(o) }
func (o *dbServiceListOutput) toTable() { outputTable(o) }

type dbServiceListCmd struct {
	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *dbServiceListCmd) cmdAliases() []string { return gListAlias }

func (c *dbServiceListCmd) cmdShort() string { return "List Database Services" }

func (c *dbServiceListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Database Services.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&dbServiceListItemOutput{}), ", "))
}

func (c *dbServiceListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbServiceListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	out := make(dbServiceListOutput, 0)
	res := make(chan dbServiceListItemOutput)
	defer close(res)

	go func() {
		for dbService := range res {
			out = append(out, dbService)
		}
	}()
	err := forEachZone(zones, func(zone string) error {
		list, err := cs.ListDatabaseServices(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Database Services in zone %s: %v", zone, err)
		}

		for _, dbService := range list {
			res <- dbServiceListItemOutput{
				Name: dbService.Name,
				Type: dbService.Type,
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

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbCmd, &dbServiceListCmd{}))
}
