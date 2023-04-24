package cmd

import (
	"fmt"
	"time"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type runstatusMaintenanceShowOutput struct {
	ID               int        `json:"id"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	State            string     `json:"state"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	AffectedServices []string   `json:"affected_services,omitempty"`
}

func (o *runstatusMaintenanceShowOutput) Type() string { return "Maintenance" }
func (o *runstatusMaintenanceShowOutput) toJSON()      { output.JSON(o) }
func (o *runstatusMaintenanceShowOutput) toText()      { output.Text(o) }
func (o *runstatusMaintenanceShowOutput) toTable()     { output.Table(o) }

func init() {
	runstatusMaintenanceCmd.AddCommand(
		&cobra.Command{
			Use:     "show [PAGE] MAINTENANCE-NAME|ID",
			Short:   "Show a maintenance details",
			Aliases: gShowAlias,
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					return cmd.Usage()
				}

				page := gCurrentAccount.DefaultRunstatusPage
				maintenance := args[0]

				if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
					return fmt.Errorf("No default runstat.us page is set.\n"+
						"Please specify a page in parameter or add it to your configuration file %s",
						gConfigFilePath)
				}

				if len(args) > 1 {
					page = args[0]
					maintenance = args[1]
				}

				return printOutput(showRunstatusMaintenance(page, maintenance))
			},
		})
}

func showRunstatusMaintenance(p, m string) (outputter, error) {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: p})
	if err != nil {
		return nil, err
	}

	maintenance, err := getRunstatusMaintenanceByNameOrID(*page, m)
	if err != nil {
		return nil, err
	}

	out := runstatusMaintenanceShowOutput{
		ID:               maintenance.ID,
		Title:            maintenance.Title,
		Description:      maintenance.Description,
		State:            maintenance.Status,
		StartDate:        maintenance.StartDate,
		EndDate:          maintenance.EndDate,
		AffectedServices: maintenance.Services,
	}

	return &out, nil
}
