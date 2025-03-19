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
		ssl := settings["ssl"].(bool)
		databaseService.ValkeySettings = &v3.JSONSchemaValkey{
			AclChannelsDefault:            v3.JSONSchemaValkeyAclChannelsDefault(settings["acl_channels_default"].(string)),
			IoThreads:                     settings["io_threads"].(int),
			LfuDecayTime:                  settings["lfu_decay_time"].(int),
			LfuLogFactor:                  settings["lfu_log_factor"].(int),
			MaxmemoryPolicy:               v3.JSONSchemaValkeyMaxmemoryPolicy(settings["maxmemory_policy"].(string)),
			NotifyKeyspaceEvents:          settings["notify_keyspace_events"].(string),
			NumberOfDatabases:             settings["number_of_databases"].(int),
			Persistence:                   v3.JSONSchemaValkeyPersistence(settings["persistence"].(string)),
			PubsubClientOutputBufferLimit: settings["pubsub_client_output_buffer_limit"].(int),
			SSL:                           &ssl,
			Timeout:                       settings["timeout"].(int),
		}
	}

	if c.ValkeyMigrationHost != "" {
		databaseService.Migration = &v3.CreateDBAASServiceValkeyRequestMigration {
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
