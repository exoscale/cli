package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbTypeListItemOutput struct {
	Name          string `json:"name"`
	LatestVersion string `json:"latest_version"`
}

type dbTypeListOutput []dbTypeListItemOutput

func (o *dbTypeListOutput) toJSON()  { outputJSON(o) }
func (o *dbTypeListOutput) toText()  { outputText(o) }
func (o *dbTypeListOutput) toTable() { outputTable(o) }

type dbTypeListCmd struct {
	_ bool `cli-cmd:"list"`
}

func (c *dbTypeListCmd) cmdAliases() []string { return nil }

func (c *dbTypeListCmd) cmdShort() string { return "List Database Service types" }

func (c *dbTypeListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Database Service types.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&dbTypeListItemOutput{}), ", "))
}

func (c *dbTypeListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbTypeListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	dbTypes, err := cs.ListDatabaseServiceTypes(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := make(dbTypeListOutput, 0)

	for _, t := range dbTypes {
		out = append(out, dbTypeListItemOutput{
			Name:          *t.Name,
			LatestVersion: *t.LatestVersion,
		})
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbTypeCmd, &dbTypeListCmd{}))
}
