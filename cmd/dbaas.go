package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/mitchellh/go-wordwrap"
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

var dbaasCmd = &cobra.Command{
	Use:   "dbaas",
	Short: "Database as a Service management",
}

func init() {
	RootCmd.AddCommand(dbaasCmd)
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

// redactDatabaseServiceURI returns a redacted version of the URI provided
// (i.e. masks potential password information).
func redactDatabaseServiceURI(u string) string {
	if uri, err := url.Parse(u); err == nil {
		return uri.Redacted()
	}

	return u
}

// dbaasShowSettings outputs a table-formatted list of key/value settings.
func dbaasShowSettings(settings map[string]interface{}) {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"key", "type", "description"})

	for k, v := range settings {
		s, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		row := []string{k}

		typ := "-"
		if v, ok := s["type"]; ok {
			typ = fmt.Sprint(v)
		}
		row = append(row, typ)

		var description string
		if v, ok := s["description"]; ok {
			description = wordwrap.WrapString(v.(string), 50)

			if v, ok := s["enum"]; ok {
				description = fmt.Sprintf("%s\n  * Supported values:\n%s", description, func() string {
					values := make([]string, len(v.([]interface{})))
					for i, val := range v.([]interface{}) {
						values[i] = fmt.Sprintf("    - %v", val)
					}
					return strings.Join(values, "\n")
				}())
			}

			min, hasMin := s["minimum"]
			max, hasMax := s["maximum"]
			if hasMin && hasMax {
				description = fmt.Sprintf("%s\n  * Minimum: %v / Maximum: %v", description, min, max)
			}

			if v, ok := s["default"]; ok {
				description = fmt.Sprintf("%s\n  * Default: %v", description, v)
			}

			if v, ok := s["example"]; ok {
				description = fmt.Sprintf("%s\n  * Example: %v", description, v)
			}
		}
		row = append(row, description)

		t.Append(row)
	}
}
