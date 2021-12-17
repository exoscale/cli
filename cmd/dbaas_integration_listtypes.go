package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasIntegrationTypeListTypesItemOutput struct {
	Type                   string   `json:"type"`
	Destinations           []string `json:"destinations"`
	DestinationDescription string   `json:"destination_description"`
	Sources                []string `json:"sources"`
	SourceDescription      string   `json:"source_description"`
}

type dbaasIntegrationListTypesOutput []dbaasIntegrationTypeListTypesItemOutput

func (o *dbaasIntegrationListTypesOutput) toJSON() { outputJSON(o) }
func (o *dbaasIntegrationListTypesOutput) toText() { outputText(o) }
func (o *dbaasIntegrationListTypesOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Type", "Sources", "Destinations"})
	defer t.Render()

	for _, i := range *o {
		t.Append([]string{
			i.Type,
			fmt.Sprintf(
				"%s (%s)",
				strings.Join(i.Sources, ", "),
				i.SourceDescription,
			),
			fmt.Sprintf(
				"%s (%s)",
				strings.Join(i.Destinations, ", "),
				i.DestinationDescription,
			),
		})
	}
}

type dbaasIntegrationListTypesCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list-types"`
}

func (c *dbaasIntegrationListTypesCmd) cmdAliases() []string { return nil }

func (c *dbaasIntegrationListTypesCmd) cmdShort() string {
	return "List Database Service integration types"
}

func (c *dbaasIntegrationListTypesCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Database Service integration types.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&dbaasIntegrationTypeListTypesItemOutput{}), ", "))
}

func (c *dbaasIntegrationListTypesCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasIntegrationListTypesCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	res, err := cs.ListDbaasIntegrationTypesWithResponse(ctx)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	out := make(dbaasIntegrationListTypesOutput, 0)

	if res.JSON200 != nil && res.JSON200.DbaasIntegrationTypes != nil {
		for _, i := range *res.JSON200.DbaasIntegrationTypes {
			out = append(out, dbaasIntegrationTypeListTypesItemOutput{
				Type:              *i.Type,
				SourceDescription: defaultString(i.SourceDescription, ""),
				Sources: func() (v []string) {
					if i.SourceServiceTypes != nil {
						v = *i.SourceServiceTypes
					}
					return
				}(),
				DestinationDescription: defaultString(i.DestDescription, ""),
				Destinations: func() (v []string) {
					if i.DestServiceTypes != nil {
						v = *i.DestServiceTypes
					}
					return
				}(),
			})
		}
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasIntegrationCmd, &dbaasIntegrationListTypesCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
