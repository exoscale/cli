package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/exoscale/cli/pkg/output"
	"github.com/spf13/cobra"
)

type runstatusIncidentListItemOutput struct {
	ID        int        `json:"id"`
	Page      string     `json:"page"`
	Title     string     `json:"title"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	State     string     `json:"state"`
	Status    string     `json:"status"`
}

type runstatusIncidentListOutput []runstatusIncidentListItemOutput

func (o *runstatusIncidentListOutput) ToJSON() { output.JSON(o) }

func (o *runstatusIncidentListOutput) ToText() { output.Text(o) }

func (o *runstatusIncidentListOutput) ToTable() {
	for i := range *o {
		(*o)[i].State = strings.ToUpper(strings.Replace((*o)[i].State, "_", " ", -1))
	}

	output.Table(o)
}

func init() {
	runstatusIncidentCmd.AddCommand(&cobra.Command{
		Use:   "list [PAGE]...",
		Short: "List incidents",
		Long: fmt.Sprintf(`This command lists existing runstat.us incidents.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&runstatusIncidentListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(runstatusListIncidents(args))
		},
	})
}

func runstatusListIncidents(pageNames []string) (output.Outputter, error) {
	pages, err := getRunstatusPages(pageNames)
	if err != nil {
		return nil, err
	}

	out := runstatusIncidentListOutput{}

	for _, page := range pages {
		incidents, err := csRunstatus.ListRunstatusIncidents(gContext, page)
		if err != nil {
			return nil, err
		}

		for _, incident := range incidents {
			out = append(out, runstatusIncidentListItemOutput{
				ID:        incident.ID,
				Title:     incident.Title,
				StartDate: incident.StartDate,
				EndDate:   incident.EndDate,
				State:     incident.State,
				Status:    incident.Status,
				Page:      page.Subdomain,
			})
		}
	}

	return &out, nil
}
