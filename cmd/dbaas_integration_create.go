package cmd

import (
	"fmt"
	"net/http"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

type dbaasIntegrationCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Type        string `cli-arg:"#"`
	Source      string `cli-arg:"#" cli-usage:"SOURCE-SERVICE"`
	Destination string `cli-arg:"#" cli-usage:"DESTINATION-SERVICE"`

	Settings string `cli-flag:"settings" cli-usage:"configuration settings (JSON format)"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasIntegrationCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *dbaasIntegrationCreateCmd) cmdShort() string { return "Create a Database Service integration" }

func (c *dbaasIntegrationCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Database Service integration.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&dbaasIntegrationShowOutput{}), ", "))
}

func (c *dbaasIntegrationCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasIntegrationCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	var integrationSettings map[string]interface{}
	if c.Settings != "" {
		// Inferring supported integration settings based on targeted source/destination services.
		databaseServices, err := cs.ListDatabaseServices(ctx, c.Zone)
		if err != nil {
			return err
		}

		var sourceService, destinationService *egoscale.DatabaseService
		for _, s := range databaseServices {
			if *s.Name == c.Source {
				sourceService = s
				continue
			}
			if *s.Name == c.Destination {
				destinationService = s
				continue
			}
		}
		if sourceService == nil {
			return fmt.Errorf("source Database Service %q not found", c.Source)
		}
		if destinationService == nil {
			return fmt.Errorf("destination Database Service %q not found", c.Destination)
		}

		settingsSchema, err := cs.ListDbaasIntegrationSettingsWithResponse(
			ctx,
			c.Type,
			*sourceService.Type,
			*destinationService.Type,
		)
		if err != nil {
			return err
		}
		if settingsSchema.StatusCode() != http.StatusOK {
			return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
		}

		settings, err := validateDatabaseServiceSettings(c.Settings, settingsSchema.JSON200.Settings)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		integrationSettings = settings
	}

	var (
		res *oapi.CreateDbaasIntegrationResponse
		err error
	)
	decorateAsyncOperation("Creating Database Service integration...", func() {
		res, err = cs.CreateDbaasIntegrationWithResponse(ctx, oapi.CreateDbaasIntegrationJSONRequestBody{
			DestService:     oapi.DbaasServiceName(c.Destination),
			IntegrationType: c.Type,
			Settings:        &integrationSettings,
			SourceService:   oapi.DbaasServiceName(c.Source),
		})
	})
	if err != nil {
		return err
	}
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	return (&dbaasIntegrationShowCmd{
		cliCommandSettings: c.cliCommandSettings,
		ID:                 *res.JSON200.Id,
		Zone:               c.Zone,
	}).cmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasIntegrationCmd, &dbaasIntegrationCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
