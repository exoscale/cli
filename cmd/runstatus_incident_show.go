package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type runstatusIncidentEventShowOutput struct {
	Date   *time.Time `json:"date"`
	Status string     `json:"status"`
	Text   string     `json:"text"`
}

type runstatusIncidentShowOutput struct {
	ID               int                                `json:"id"`
	Title            string                             `json:"title"`
	StartDate        *time.Time                         `json:"start_date"`
	EndDate          *time.Time                         `json:"end_date"`
	State            string                             `json:"state"`
	Status           string                             `json:"status"`
	AffectedServices []string                           `json:"affected_services,omitempty"`
	Events           []runstatusIncidentEventShowOutput `json:"events,omitempty"`
}

func (o *runstatusIncidentShowOutput) toJSON() { outputJSON(o) }

func (o *runstatusIncidentShowOutput) toText() { outputText(o) }

func (o *runstatusIncidentShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Incident"})

	t.Append([]string{"ID", fmt.Sprint(o.ID)})
	t.Append([]string{"Title", o.Title})
	t.Append([]string{"State", o.State})
	t.Append([]string{"Status", o.Status})
	t.Append([]string{"Start Date", fmt.Sprint(o.StartDate)})

	if o.EndDate != nil {
		t.Append([]string{"End Date", fmt.Sprint(o.EndDate)})
	}

	t.Append([]string{"Affected Services", strings.Join(o.AffectedServices, "\n")})

	if len(o.Events) > 0 {
		buf := bytes.NewBuffer(nil)
		et := table.NewEmbeddedTable(buf)
		et.SetHeader([]string{" "})
		for _, e := range o.Events {
			et.Append([]string{fmt.Sprint(e.Date), e.Status, e.Text})
		}
		et.Render()
		t.Append([]string{"Event Stream", buf.String()})
	}

	t.Render()
}

func init() {
	runstatusIncidentCmd.AddCommand(&cobra.Command{
		Use:   "show [page name] <incident name | id>",
		Short: "Show an incident details",
		Long: fmt.Sprintf(`This command shows a runstat.us incident details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&runstatusIncidentShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			page := gCurrentAccount.DefaultRunstatusPage
			incident := args[0]

			if gCurrentAccount.DefaultRunstatusPage == "" && len(args) == 1 {
				return fmt.Errorf("No default runstat.us page is set.\n"+
					"Please specify a page in parameter or add it to your configuration file %s",
					gConfigFilePath)
			}

			if len(args) > 1 {
				page = args[0]
				incident = args[1]
			}

			return output(showRunstatusIncident(page, incident))
		},
	})
}

func showRunstatusIncident(p, i string) (outputter, error) {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: p})
	if err != nil {
		return nil, err
	}

	incident, err := getIncidentByNameOrID(*page, i)
	if err != nil {
		return nil, err
	}

	out := runstatusIncidentShowOutput{
		ID:               incident.ID,
		Title:            incident.Title,
		StartDate:        incident.StartDate,
		EndDate:          incident.EndDate,
		State:            incident.State,
		Status:           incident.Status,
		AffectedServices: incident.Services,
	}

	if n := len(incident.Events); n > 0 {
		out.Events = make([]runstatusIncidentEventShowOutput, n)
		for i, e := range incident.Events {
			out.Events[i] = runstatusIncidentEventShowOutput{
				Date:   e.Created,
				Status: e.Status,
				Text:   e.Text}
		}
	}

	return &out, nil
}
