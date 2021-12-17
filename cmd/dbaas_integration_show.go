package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasIntegrationShowOutput struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Enabled     bool   `json:"enabled"`
	Active      bool   `json:"active"`
	Status      string `json:"status"`
}

func (o *dbaasIntegrationShowOutput) toJSON()  { outputJSON(o) }
func (o *dbaasIntegrationShowOutput) toText()  { outputText(o) }
func (o *dbaasIntegrationShowOutput) toTable() { outputTable(o) }

type dbaasIntegrationShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ID string `cli-arg:"#" cli-usage:"UUID"`

	ShowSettings bool   `cli-flag:"settings" cli-usage:""`
	Zone         string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasIntegrationShowCmd) cmdAliases() []string { return gShowAlias }

func (c *dbaasIntegrationShowCmd) cmdShort() string {
	return "Show a Database Service integration details"
}

func (c *dbaasIntegrationShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service integration details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&dbaasIntegrationShowOutput{}), ", "))
}

func (c *dbaasIntegrationShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasIntegrationShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	res, err := cs.GetDbaasIntegrationWithResponse(ctx, c.ID)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}
	databaseIntegration := res.JSON200

	if c.ShowSettings {
		if databaseIntegration.Settings != nil {
			out, err := json.MarshalIndent(databaseIntegration.Settings, "", "  ")
			if err != nil {
				return fmt.Errorf("unable to marshal JSON: %w", err)
			}
			fmt.Println(string(out))
		}

		return nil
	}

	return c.outputFunc(&dbaasIntegrationShowOutput{
		ID:          *databaseIntegration.Id,
		Type:        *databaseIntegration.Type,
		Description: defaultString(databaseIntegration.Description, ""),
		Source:      *databaseIntegration.Source,
		Destination: *databaseIntegration.Dest,
		Enabled:     *databaseIntegration.IsEnabled,
		Active:      *databaseIntegration.IsActive,
		Status:      *databaseIntegration.Status,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasIntegrationCmd, &dbaasIntegrationShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
