package servicetype

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/table"
)

var Cmd = &cobra.Command{
	Use:   "type",
	Short: "Database Services types management",
}

// Settings groups supported by each Database Service type.
var (
	GrafanaSettings    = []string{"grafana"}
	OpensearchSettings = []string{"opensearch"}
	KafkaSettings      = []string{
		"kafka",
		"kafka-rest",
		"kafka-connect",
		"schema-registry",
	}
	MysqlSettings = []string{"mysql"}
	PGSettings    = []string{
		"pg",
		"pgbouncer",
		"pglookout",
		"timescaledb",
	}
	ValkeySettings     = []string{"valkey"}
	ThanosSettings     = []string{"thanos", "compactor", "query", "query-frontend"}
	ClickhouseSettings = []string{"clickhouse"}
)

// showSettings outputs a table-formatted list of key/value settings.
func showSettings(settings map[string]any) {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"key", "type", "description"})

	for k, v := range settings {
		s, ok := v.(map[string]any)
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
					values := make([]string, len(v.([]any)))
					for i, val := range v.([]any) {
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
