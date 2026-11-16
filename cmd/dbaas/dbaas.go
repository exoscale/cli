package dbaas

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"

	exocmd "github.com/exoscale/cli/cmd"

	"github.com/exoscale/cli/cmd/dbaas/acl"
	"github.com/exoscale/cli/cmd/dbaas/database"
	externalendpoint "github.com/exoscale/cli/cmd/dbaas/external-endpoint"
	externalintegration "github.com/exoscale/cli/cmd/dbaas/external-integration"
	"github.com/exoscale/cli/cmd/dbaas/migration"
	readreplica "github.com/exoscale/cli/cmd/dbaas/read-replica"
	"github.com/exoscale/cli/cmd/dbaas/role"
	servicetype "github.com/exoscale/cli/cmd/dbaas/type"
	"github.com/exoscale/cli/cmd/dbaas/user"
)

var dbServiceMaintenanceDOWs = []string{
	"never",
	"monday",
	"tuesday",
	"wednesday",
	"thursday",
	"friday",
	"saturday",
	"sunday",
}

var dbaasCmd = &cobra.Command{
	Use:   "dbaas",
	Short: "Database as a Service management",
}

func init() {
	exocmd.RootCmd.AddCommand(dbaasCmd)
	// Attach subcommand groups that live in their own packages.
	dbaasCmd.AddCommand(role.Cmd)
	dbaasCmd.AddCommand(acl.Cmd)
	dbaasCmd.AddCommand(database.Cmd)
	dbaasCmd.AddCommand(externalendpoint.Cmd)
	dbaasCmd.AddCommand(externalintegration.Cmd)
	dbaasCmd.AddCommand(migration.Cmd)
	dbaasCmd.AddCommand(readreplica.Cmd)
	dbaasCmd.AddCommand(servicetype.Cmd)
	dbaasCmd.AddCommand(user.Cmd)
}

// parseDtabaseBackupSchedule parses a Database Service backup schedule value
// expressed in HH:MM format and returns the discrete values for hour and
// minute, or an error if the parsing failed.
func parseDatabaseBackupSchedule(v string) (int64, int64, error) {
	parts := strings.Split(v, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid value %q for backup schedule, expecting HH:MM", v)
	}

	backupHour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid value %q for backup schedule hour, must be between 0 and 23", v)
	}

	backupMinute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid value %q for backup schedule minute, must be between 0 and 59", v)
	}

	return int64(backupHour), int64(backupMinute), nil
}

// validateDatabaseServiceSettings validates user-provided JSON-formatted
// Database Service settings against a reference JSON Schema.
func validateDatabaseServiceSettings(in string, schema interface{}) (map[string]interface{}, error) {
	var userSettings map[string]interface{}

	if err := json.Unmarshal([]byte(in), &userSettings); err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON: %w", err)
	}

	res, err := gojsonschema.Validate(
		gojsonschema.NewGoLoader(schema),
		gojsonschema.NewStringLoader(in),
	)
	if err != nil {
		// JSON Schema is provided by API and if loading fails there is nothing a user can to to fix the issue.
		// One example is incompatible regex engines for pattern validation that will prevent loading JSON schema.
		// When that happens we should still allow running the command as API would validate request.
		return userSettings, nil
	}

	if !res.Valid() {
		for _, err := range res.Errors() {
			errs := []string{}
			// Some regexs are known not to match in Go (they are written for Python).
			// Thus we ignore pattern errors and rely on server side validation for them.
			if err.Type() != "pattern" {
				errs = append(errs, err.String())
			}

			if len(errs) > 0 {
				return nil, errors.New(strings.Join(errs, "\n"))
			}
		}
	}

	return userSettings, nil
}
