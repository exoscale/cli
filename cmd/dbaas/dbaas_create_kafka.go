package dbaas

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceCreateCmd) createKafka(_ *cobra.Command, _ []string) error {
	var err error

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.CreateDBAASServiceKafkaRequest{
		KafkaConnectEnabled:   &c.KafkaEnableKafkaConnect,
		KafkaRestEnabled:      &c.KafkaEnableKafkaREST,
		Plan:                  c.Plan,
		SchemaRegistryEnabled: &c.KafkaEnableSchemaRegistry,
		TerminationProtection: &c.TerminationProtection,
	}

	if c.KafkaVersion != "" {
		databaseService.Version = c.KafkaVersion
	}

	settingsSchema, err := client.GetDBAASSettingsKafka(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}

	if c.KafkaEnableCertAuth || c.KafkaEnableSASLAuth {
		databaseService.AuthenticationMethods = &v3.CreateDBAASServiceKafkaRequestAuthenticationMethods{
			Certificate: &c.KafkaEnableCertAuth,
			Sasl:        &c.KafkaEnableSASLAuth,
		}
	}

	if len(c.KafkaIPFilter) > 0 {
		databaseService.IPFilter = c.KafkaIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServiceKafkaRequestMaintenance{
			Dow:  v3.CreateDBAASServiceKafkaRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.KafkaConnectSettings != "" {
		_, err := validateDatabaseServiceSettings(
			c.KafkaConnectSettings,
			settingsSchema.Settings.KafkaConnect,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		settings := &v3.JSONSchemaKafkaConnect{}
		if err = json.Unmarshal([]byte(c.KafkaConnectSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		databaseService.KafkaConnectSettings = settings
	}

	if c.KafkaRESTSettings != "" {
		_, err := validateDatabaseServiceSettings(
			c.KafkaRESTSettings,
			settingsSchema.Settings.KafkaRest,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaKafkaRest{}
		if err = json.Unmarshal([]byte(c.KafkaRESTSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.KafkaRestSettings = settings
	}

	if c.KafkaSettings != "" {
		_, err := validateDatabaseServiceSettings(
			c.KafkaSettings,
			settingsSchema.Settings.Kafka,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaKafka{}
		if err = json.Unmarshal([]byte(c.KafkaSettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.KafkaSettings = settings
	}

	if c.KafkaSchemaRegistrySettings != "" {
		_, err := validateDatabaseServiceSettings(
			c.KafkaSchemaRegistrySettings,
			settingsSchema.Settings.SchemaRegistry,
		)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		settings := &v3.JSONSchemaSchemaRegistry{}
		if err = json.Unmarshal([]byte(c.KafkaSchemaRegistrySettings), settings); err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}
		databaseService.SchemaRegistrySettings = settings
	}

	op, err := client.CreateDBAASServiceKafka(ctx, c.Name, databaseService)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Database Service %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceKafka(ctx))
	}

	return nil
}
