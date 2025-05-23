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

func (c *dbaasServiceUpdateCmd) updateValkey(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := GContext

	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}

	databaseService := v3.UpdateDBAASServiceValkeyRequest{}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.ValkeyIPFilter)) {
		databaseService.IPFilter = c.ValkeyIPFilter
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
		databaseService.Maintenance = &v3.UpdateDBAASServiceValkeyRequestMaintenance{
			Dow:  v3.UpdateDBAASServiceValkeyRequestMaintenanceDow(c.MaintenanceDOW),
			Time: c.MaintenanceTime,
		}
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.ValkeySettings)) {
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
		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.ValkeyMigrationHost)) {
		databaseService.Migration = &v3.UpdateDBAASServiceValkeyRequestMigration{
			Host: c.ValkeyMigrationHost,
			Port: c.ValkeyMigrationPort,
		}
		if c.ValkeyMigrationPassword != "" {
			databaseService.Migration.Password = c.ValkeyMigrationPassword
		}
		if c.ValkeyMigrationUsername != "" {
			databaseService.Migration.Username = c.ValkeyMigrationUsername
		}
		if c.ValkeyMigrationDBName != "" {
			databaseService.Migration.Dbname = c.ValkeyMigrationDBName
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
		updated = true
	}

	if updated {
		op, err := client.UpdateDBAASServiceValkey(ctx, c.Name, databaseService)
		if err != nil {
			return err
		}

		decorateAsyncOperation(fmt.Sprintf("Updating DBaaS Datadog external Endpoint %q", c.Name), func() {
			op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})

		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return c.OutputFunc((&dbaasServiceShowCmd{
			Name: c.Name,
			Zone: c.Zone,
		}).showDatabaseServiceValkey(ctx))
	}
	return nil
}
