package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type runstatusMaintenanceListItemOutput struct {
	ID        int        `json:"id"`
	Page      string     `json:"page"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type runstatusMaintenanceListOutput []runstatusMaintenanceListItemOutput

func (o *runstatusMaintenanceListOutput) toJSON()  { outputJSON(o) }
func (o *runstatusMaintenanceListOutput) toText()  { outputText(o) }
func (o *runstatusMaintenanceListOutput) toTable() { outputTable(o) }

func init() {
	runstatusMaintenanceCmd.AddCommand(&cobra.Command{
		Use:   "list [page name ...]",
		Short: "List maintenance from page(s)",
		Long: fmt.Sprintf(`This command lists existing runstat.us maintenances.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&runstatusMaintenanceListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(runstatusListMaintenances(args))
		},
	})
}

func runstatusListMaintenances(pageNames []string) (outputter, error) {
	pages, err := getRunstatusPages(pageNames)
	if err != nil {
		return nil, err
	}

	out := runstatusMaintenanceListOutput{}

	for _, page := range pages {
		maintenances, err := csRunstatus.ListRunstatusMaintenances(gContext, page)
		if err != nil {
			return nil, err
		}

		for _, maintenance := range maintenances {
			out = append(out, runstatusMaintenanceListItemOutput{
				ID:        maintenance.ID,
				Title:     maintenance.Title,
				StartDate: maintenance.StartDate,
				EndDate:   maintenance.EndDate,
				Status:    maintenance.Status,
				Page:      page.Subdomain,
			})
		}
	}

	return &out, nil
}
