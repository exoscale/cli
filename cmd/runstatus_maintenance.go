package cmd

import (
	"strconv"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusMaintenanceCmd represents the maintenance command
var runstatusMaintenanceCmd = &cobra.Command{
	Use:   "maintenance",
	Short: "Maintenance management",
}

func getMaintenanceByNameOrID(page egoscale.RunstatusPage, name string) (*egoscale.RunstatusMaintenance, error) {
	m := egoscale.RunstatusMaintenance{PageURL: page.URL}

	if id, err := strconv.Atoi(name); err == nil {
		m.ID = id
	} else {
		m.Title = name
	}

	return csRunstatus.GetRunstatusMaintenance(gContext, m)
}

func init() {
	runstatusCmd.AddCommand(runstatusMaintenanceCmd)
}
