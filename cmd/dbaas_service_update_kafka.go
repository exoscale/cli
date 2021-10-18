package cmd

import (
	"fmt"
	"net/http"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

func (c *dbServiceUpdateCmd) updateKafka(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	databaseService := oapi.UpdateDbaasServiceKafkaJSONRequestBody{}

	settingsSchema, err := cs.GetDbaasSettingsKafkaWithResponse(ctx)
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

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceKafkaJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{}
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) {
			databaseService.Maintenance.Dow = oapi.UpdateDbaasServiceKafkaJSONBodyMaintenanceDow(c.MaintenanceDOW)
		}
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
			databaseService.Maintenance.Time = c.MaintenanceTime
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
		fmt.Printf("Updating Database Service %q...\n", c.Name)

		res, err := cs.UpdateDbaasServiceKafkaWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		if err != nil {
			return err
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("API request error: unexpected status %s", res.Status())
		}
	}

	if !gQuiet {
		return output((&dbServiceShowCmd{Zone: c.Zone, Name: c.Name}).showDatabaseServiceKafka(ctx))
	}

	return nil
}
