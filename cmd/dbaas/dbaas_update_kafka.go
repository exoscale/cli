package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	v3 "github.com/exoscale/egoscale/v3"
=======
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
)

func (c *dbaasServiceUpdateCmd) updateKafka(cmd *cobra.Command, _ []string) error {
	var updated bool

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	ctx := gContext
=======
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	databaseService := v3.UpdateDBAASServiceKafkaRequest{}

	settingsSchema, err := client.GetDBAASSettingsKafka(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve database Service settings: %w", err)
	}

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) ||
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableSASLAuth)) {
		databaseService.AuthenticationMethods = &v3.UpdateDBAASServiceKafkaRequestAuthenticationMethods{}
		if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) {
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) ||
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaEnableSASLAuth)) {
		databaseService.AuthenticationMethods = &struct {
			Certificate *bool `json:"certificate,omitempty"`
			Sasl        *bool `json:"sasl,omitempty"`
		}{}
		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaEnableCertAuth)) {
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
			databaseService.AuthenticationMethods.Certificate = &c.KafkaEnableCertAuth
		}
		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaEnableSASLAuth)) {
			databaseService.AuthenticationMethods.Sasl = &c.KafkaEnableSASLAuth
		}
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaEnableKafkaConnect)) {
		databaseService.KafkaConnectEnabled = &c.KafkaEnableKafkaConnect
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaEnableKafkaREST)) {
		databaseService.KafkaRestEnabled = &c.KafkaEnableKafkaREST
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaEnableSchemaRegistry)) {
		databaseService.SchemaRegistryEnabled = &c.KafkaEnableSchemaRegistry
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaIPFilter)) {
		databaseService.IPFilter = c.KafkaIPFilter
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = c.Plan
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaIPFilter)) {
		databaseService.IpFilter = &c.KafkaIPFilter
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Plan)) {
		databaseService.Plan = &c.Plan
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.TerminationProtection)) {
		databaseService.TerminationProtection = &c.TerminationProtection
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(mustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &v3.UpdateDBAASServiceKafkaRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceKafkaRequestMaintenanceDow(c.MaintenanceDOW),
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceDOW)) &&
		cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.MaintenanceTime)) {
		databaseService.Maintenance = &struct {
			Dow  oapi.UpdateDbaasServiceKafkaJSONBodyMaintenanceDow `json:"dow"`
			Time string                                             `json:"time"`
		}{
			Dow:  oapi.UpdateDbaasServiceKafkaJSONBodyMaintenanceDow(c.MaintenanceDOW),
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
			Time: c.MaintenanceTime,
		}
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaConnectSettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaConnectSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
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

		databaseService.KafkaConnectSettings = *settings
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaRESTSettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaRESTSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
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
		databaseService.KafkaRestSettings = *settings
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaSettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaSettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
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
		databaseService.KafkaSettings = *settings
		updated = true
	}

<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.KafkaSchemaRegistrySettings)) {
		_, err := validateDatabaseServiceSettings(
=======
	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.KafkaSchemaRegistrySettings)) {
		settings, err := validateDatabaseServiceSettings(
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
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
		databaseService.SchemaRegistrySettings = *settings
		updated = true
	}

	if updated {
<<<<<<< Updated upstream:cmd/dbaas_update_kafka.go
		op, err := client.UpdateDBAASServiceKafka(ctx, c.Name, databaseService)
=======
		var res *oapi.UpdateDbaasServiceKafkaResponse
		utils.DecorateAsyncOperation(fmt.Sprintf("Updating Database Service %q...", c.Name), func() {
			res, err = globalstate.EgoscaleClient.UpdateDbaasServiceKafkaWithResponse(ctx, oapi.DbaasServiceName(c.Name), databaseService)
		})
>>>>>>> Stashed changes:cmd/dbaas/dbaas_update_kafka.go
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
