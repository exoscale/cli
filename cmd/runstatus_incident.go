package cmd

import (
	"strconv"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var runstatusIncidentCmd = &cobra.Command{
	Use:   "incident",
	Short: "Incident management",
}

func getRunstatusIncidentByNameOrID(page egoscale.RunstatusPage, name string) (*egoscale.RunstatusIncident, error) {
	i := egoscale.RunstatusIncident{PageURL: page.URL}

	if id, err := strconv.Atoi(name); err == nil {
		i.ID = id
	} else {
		i.Title = name
	}

	return csRunstatus.GetRunstatusIncident(gContext, i)
}

func init() {
	runstatusCmd.AddCommand(runstatusIncidentCmd)
}
