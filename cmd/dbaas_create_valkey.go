package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasServiceCreateCmd) createValkey(_ *cobra.Command, _ []string) error {
	var err error

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))

	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.CreateDBAASServiceValkeyRequest{
		Plan:                  c.Plan,
		TerminationProtection: &c.TerminationProtection,
	}

	settingsSchema, err := client.GetDBAASSettingsValkey(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve Database Service settings: %w", err)
	}

	if c.ValkeyForkFrom != "" {
		databaseService.ForkFromService = v3.DBAASServiceName(c.ValkeyForkFrom)
		if c.ValkeyRecoveryBackupName != "" {
			databaseService.RecoveryBackupName = c.ValkeyRecoveryBackupName
		}
	}

	if len(c.ValkeyIPFilter) > 0 {
		databaseService.IPFilter = c.ValkeyIPFilter
	}

	if c.MaintenanceDOW != "" && c.MaintenanceTime != "" {
		databaseService.Maintenance = &v3.CreateDBAASServiceValkeyRequestMaintenance{
			Dow:  v3.CreateDBAASServiceValkeyRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
	}

	if c.ValkeySettings != "" {
		settings, err := validateDatabaseServiceSettings(c.ValkeySettings, settingsSchema.Settings.Valkey)
		if err != nil {
			return fmt.Errorf("invalid settings: %w", err)
		}

		valkeysettings := &v3.JSONSchemaValkey{}

		if val, ok := settings["acl_channels_default"]; ok && val != nil {
			valkeysettings.AclChannelsDefault = v3.JSONSchemaValkeyAclChannelsDefault(val.(string))
		}
		if val, ok := settings["io_threads"]; ok && val != nil {
			valkeysettings.IoThreads = int(val.(float64))
		}
		if val, ok := settings["lfu_decay_time"]; ok && val != nil {
			valkeysettings.LfuDecayTime = int(val.(float64))
		}
		if val, ok := settings["lfu_log_factor"]; ok && val != nil {
			valkeysettings.LfuLogFactor = int(val.(float64))
		}
		if val, ok := settings["maxmemory_policy"]; ok && val != nil {
			valkeysettings.MaxmemoryPolicy = v3.JSONSchemaValkeyMaxmemoryPolicy(val.(string))
		}
		if val, ok := settings["notify_keyspace_events"]; ok && val != nil {
			valkeysettings.NotifyKeyspaceEvents = val.(string)
		}
		if val, ok := settings["number_of_databases"]; ok && val != nil {
			valkeysettings.NumberOfDatabases = int(val.(float64))
		}
		if val, ok := settings["persistence"]; ok && val != nil {
			valkeysettings.Persistence = v3.JSONSchemaValkeyPersistence(val.(string))
		}
		if val, ok := settings["pubsub_client_output_buffer_limit"]; ok && val != nil {
			valkeysettings.PubsubClientOutputBufferLimit = int(val.(float64))
		}
		if val, ok := settings["ssl"]; ok && val != nil {
			ssl := val.(bool)
			valkeysettings.SSL = &ssl
		}
		if val, ok := settings["timeout"]; ok && val != nil {
			valkeysettings.Timeout = int(val.(float64))
		}

		databaseService.ValkeySettings = valkeysettings
	}

	if c.ValkeyMigrationHost != "" {
		databaseService.Migration = &v3.CreateDBAASServiceValkeyRequestMigration{
			Host:     c.ValkeyMigrationHost,
			Port:     c.ValkeyMigrationPort,
			Password: c.ValkeyMigrationPassword,
			Username: c.ValkeyMigrationUsername,
			Dbname:   c.ValkeyMigrationDBName,
		}
		if c.ValkeyMigrationSSL {
			databaseService.Migration.SSL = &c.ValkeyMigrationSSL
		}
		if c.ValkeyMigrationMethod != "" {
			databaseService.Migration.Method = v3.EnumMigrationMethod(c.ValkeyMigrationMethod)
		}
		if len(c.ValkeyMigrationIgnoreDbs) > 0 {
			databaseService.Migration.IgnoreDbs = strings.Join(c.ValkeyMigrationIgnoreDbs, ",")
		}
	}

	op, err := client.CreateDBAASServiceValkey(ctx, c.Name, databaseService)

	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating DBaaS Datadog external Endpoint %q", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return err
	}

	serviceName := op.Reference.ID.String()

	if !globalstate.Quiet {
		return c.outputFunc((&dbaasServiceShowCmd{
			Name: serviceName,
			Zone: c.Zone,
		}).showDatabaseServiceValkey(ctx))
	}

	return nil
}
