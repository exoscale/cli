package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceUpdateCmd) updateKafka(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.UpdateDBAASServiceKafkaRequest{}

	settingsSchema, err := client.GetDBAASSettingsKafka(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) ||
		cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaEnableSASLAuth)) {
		databaseService.AuthenticationMethods = &v3.UpdateDBAASServiceKafkaRequestAuthenticationMethods{}
		if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) {
			databaseService.AuthenticationMethods.Certificate = &c.KafkaEnableCertAuth
		}
		if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaEnableSASLAuth)) {
			databaseService.AuthenticationMethods.Sasl = &c.KafkaEnableSASLAuth
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaEnableKafkaConnect)) {
		databaseService.KafkaConnectEnabled = &c.KafkaEnableKafkaConnect
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaEnableKafkaREST)) {
		databaseService.KafkaRestEnabled = &c.KafkaEnableKafkaREST
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaEnableSchemaRegistry)) {
		databaseService.SchemaRegistryEnabled = &c.KafkaEnableSchemaRegistry
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaIPFilter)) {
		databaseService.IPFilter = c.KafkaIPFilter
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServiceKafkaRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceKafkaRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaConnectSettings)) {
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
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaRESTSettings)) {
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
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaSettings)) {
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
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.KafkaSchemaRegistrySettings)) {
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
		updated = true
	}

	if updated {
		op, err := client.UpdateDBAASServiceKafka(ctx, c.Name, databaseService)
		if err != nil {
			if errors.Is(err, v3.ErrNotFound) {
				return fmt.Errorf("resource not found in zone %q", c.Zone)
			}
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})
		if err != nil {
			return err
		}

	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceKafka(ctx))
	}

	return nil
}
