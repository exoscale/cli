package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
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

var dbCmd = &cobra.Command{
	Use:   "dbaas",
	Short: "Database as a Service management",
}

func init() {
	RootCmd.AddCommand(dbCmd)
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
		return nil, fmt.Errorf("unable to validate JSON Schema: %w", err)
	}

	if !res.Valid() {
		return nil, errors.New(strings.Join(
			func() []string {
				errs := make([]string, len(res.Errors()))
				for i, err := range res.Errors() {
					errs[i] = err.String()
				}
				return errs
			}(),
			"\n",
		))
	}

	return userSettings, nil
}
