package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dbaasIntegrationShowSettingsCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show-settings"`

	Type        string `cli-arg:"#"`
	Source      string `cli-arg:"#"`
	Destination string `cli-arg:"#"`
}

func (c *dbaasIntegrationShowSettingsCmd) cmdAliases() []string { return nil }

func (c *dbaasIntegrationShowSettingsCmd) cmdShort() string {
	return "Show Database Service integration settings"
}

func (c *dbaasIntegrationShowSettingsCmd) cmdLong() string {
	return "This command shows supported settings for a specific type/source/destination Database Service integration."
}

func (c *dbaasIntegrationShowSettingsCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasIntegrationShowSettingsCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	res, err := cs.ListDbaasIntegrationSettingsWithResponse(ctx, c.Type, c.Source, c.Destination)
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	if res.JSON200 != nil && res.JSON200.Settings != nil && res.JSON200.Settings.Properties != nil {
		out, err := json.MarshalIndent(res.JSON200.Settings.Properties, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Println(string(out))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasIntegrationCmd, &dbaasIntegrationShowSettingsCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
