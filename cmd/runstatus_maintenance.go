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
	maintenanceID := -1

	id, err := strconv.Atoi(name)
	if err == nil {
		maintenanceID = id
	}

	if maintenanceID > 0 {
		return csRunstatus.GetRunstatusMaintenance(gContext, egoscale.RunstatusMaintenance{PageURL: page.URL, ID: maintenanceID})
	}

	return csRunstatus.GetRunstatusMaintenance(gContext, egoscale.RunstatusMaintenance{PageURL: page.URL, Title: name})
}

func init() {
	runstatusCmd.AddCommand(runstatusMaintenanceCmd)
}
