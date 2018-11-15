package cmd

import (
	"strconv"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// runstatusIncidentCmd represents the incident command
var runstatusIncidentCmd = &cobra.Command{
	Use:   "incident",
	Short: "Incident management",
}

func getIncidentByNameOrID(page egoscale.RunstatusPage, name string) (*egoscale.RunstatusIncident, error) {

	incidentID := -1

	id, err := strconv.Atoi(name)
	if err == nil {
		incidentID = id
	}

	if incidentID > 0 {
		return csRunstatus.GetRunstatusIncident(gContext, page, egoscale.RunstatusIncident{ID: incidentID})
	}

	return csRunstatus.GetRunstatusIncident(gContext, page, egoscale.RunstatusIncident{Title: name})
}

func init() {
	runstatusCmd.AddCommand(runstatusIncidentCmd)
}
