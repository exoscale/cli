package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	utils "github.com/exoscale/cli/utils"
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
		var settings map[string]interface{}

		if err := json.Unmarshal([]byte(c.ValkeySettings), &settings); err != nil {
			return err
		}

		ssl := utils.GetSettingBool(settings, "ssl")
		databaseService.ValkeySettings = &v3.JSONSchemaValkey{
			AclChannelsDefault:            v3.JSONSchemaValkeyAclChannelsDefault(utils.GetSettingString(settings, "acl_channels_default")),
			IoThreads:                     utils.GetSettingFloat64(settings, "io_threads"),
			LfuDecayTime:                  utils.GetSettingFloat64(settings, "lfu_decay_time"),
			LfuLogFactor:                  utils.GetSettingFloat64(settings, "lfu_log_factor"),
			MaxmemoryPolicy:               v3.JSONSchemaValkeyMaxmemoryPolicy(utils.GetSettingString(settings, "maxmemory_policy")),
			NotifyKeyspaceEvents:          utils.GetSettingString(settings, "notify_keyspace_events"),
			NumberOfDatabases:             utils.GetSettingFloat64(settings, "number_of_databases"),
			Persistence:                   v3.JSONSchemaValkeyPersistence(utils.GetSettingString(settings, "persistence")),
			PubsubClientOutputBufferLimit: utils.GetSettingFloat64(settings, "pubsub_client_output_buffer_limit"),
			SSL:                           &ssl,
			Timeout:                       utils.GetSettingFloat64(settings, "timeout"),
		}
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
