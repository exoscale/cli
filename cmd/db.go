package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

func getDatabaseServiceUserConfigFromFile(path string) (map[string]interface{}, error) {
	var userConfig map[string]interface{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &userConfig); err != nil {
		return nil, err
	}

	return userConfig, nil
}

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
	Use:     "database",
	Aliases: []string{"db"},
	Short:   "Database Services management",
}

func init() {
	labCmd.AddCommand(dbCmd)
}
