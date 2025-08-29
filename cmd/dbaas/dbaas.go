package dbaas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
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
}

// parseDtabaseBackupSchedule parses a Database Service backup schedule value
// expressed in HH:MM format and returns the discrete values for hour and
// minute, or an error if the parsing failed.
func parseDatabaseBackupSchedule(v string) (*int64, *int64, error) {
	parts := strings.Split(v, ":")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("invalid value %q for backup schedule, expecting HH:MM", v)
	}

	backupHour, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value %q for backup schedule hour, must be between 0 and 23", v)
	}

	backupMinute, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid value %q for backup schedule minute, must be between 0 and 59", v)
	}
	h, m := int64(backupHour), int64(backupMinute)
	return &h, &m, nil
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

func dbaasGetV3(ctx context.Context, name, zone string) (v3.DBAASServiceCommon, error) {

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
	if err != nil {
		return v3.DBAASServiceCommon{}, err
	}

	dbs, err := client.ListDBAASServices(ctx)
	if err != nil {
		return v3.DBAASServiceCommon{}, err
	}

	for _, db := range dbs.DBAASServices {
		if string(db.Name) == name {
			return db, nil
		}
	}

	return v3.DBAASServiceCommon{}, fmt.Errorf("%q Database Service not found in zone %q", name, zone)
}
