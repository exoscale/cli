package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

func (c *dbaasServiceUpdateCmd) updateKafka(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	databaseService := oapi.UpdateDbaasServiceKafkaJSONRequestBody{}

	settingsSchema, err := globalstate.EgoscaleClient.GetDbaasSettingsKafkaWithResponse(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}
	if settingsSchema.StatusCode() != http.StatusOK {
		return fmt.Errorf("API request error: unexpected status %s", settingsSchema.Status())
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) ||
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableSASLAuth)) {
		databaseService.AuthenticationMethods = &struct {
			Certificate *bool `json:"certificate,omitempty"`
			Sasl        *bool `json:"sasl,omitempty"`
		}{}
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) {
			databaseService.AuthenticationMethods.Certificate = &c.KafkaEnableCertAuth
		}
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableSASLAuth)) {
			databaseService.AuthenticationMethods.Sasl = &c.KafkaEnableSASLAuth
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableKafkaConnect)) {
		databaseService.KafkaConnectEnabled = &c.KafkaEnableKafkaConnect
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableKafkaREST)) {
		databaseService.KafkaRestEnabled = &c.KafkaEnableKafkaREST
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableSchemaRegistry)) {
		databaseService.SchemaRegistryEnabled = &c.KafkaEnableSchemaRegistry
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaIPFilter)) {
		databaseService.IpFilter = &c.KafkaIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceKafkaJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceKafkaJSONBodyMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaConnectSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.KafkaConnectSettings,
			settingsSchema.JSON200.Settings.KafkaConnect,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.KafkaConnectSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaRESTSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.KafkaRESTSettings,
			settingsSchema.JSON200.Settings.KafkaRest,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.KafkaRestSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaSettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.KafkaSettings,
			settingsSchema.JSON200.Settings.Kafka,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.KafkaSettings = &settings
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaSchemaRegistrySettings)) {
		settings, err := validateDatabaseServiceSettings(
			c.KafkaSchemaRegistrySettings,
			settingsSchema.JSON200.Settings.SchemaRegistry,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.SchemaRegistrySettings = &settings
		updated = true
	}

	if updated {
		var res *oapi.UpdateDbaasServiceKafkaResponse
		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceKafkaWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
		if err != nil {
			if errors.Is(err, exoapi.ErrNotFound) {
				return fmt.Errorf("resource not found in zone %q", c.Zone)
			}
			return err
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("API request error: unexpected status %s", res.Status())
		}
	}

	if !globalstate.Quiet {
		return c.outputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceKafka(ctx))
	}

	return nil
}
