package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type runstatusPageIncidentShowOutput struct {
	Title     string     `json:"title"`
	State     string     `json:"state"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type runstatusPageMaintenanceShowOutput struct {
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type runstatusPageShowOutput struct {
	ID           int                                  `json:"id"`
	Name         string                               `json:"name"`
	URL          string                               `json:"page_url"`
	Timezone     string                               `json:"timezone"`
	State        string                               `json:"state"`
	CustomDomain string                               `json:"custom_domain,omitempty"`
	Title        string                               `json:"title,omitempty"`
	SupportEmail string                               `json:"support_email,omitempty"`
	Services     map[string]string                    `json:"services,omitempty"`
	Incidents    []runstatusPageIncidentShowOutput    `json:"incidents,omitempty"`
	Maintenances []runstatusPageMaintenanceShowOutput `json:"maintenances,omitempty"`
}

func (o *runstatusPageShowOutput) toJSON() { outputJSON(o) }

func (o *runstatusPageShowOutput) toText() { outputText(o) }

func (o *runstatusPageShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Runstatus"})

	t.Append([]string{"ID", fmt.Sprint(o.ID)})
	t.Append([]string{"Name", o.Name})
	if o.Title != "" {
		t.Append([]string{"Title", o.Title})
	}
	t.Append([]string{"Page URL", o.URL})
	if o.CustomDomain != "" {
		t.Append([]string{"Custom Domain", o.CustomDomain})
	}

	t.Append([]string{"Timezone", o.Timezone})
	t.Append([]string{"State", strings.ToUpper(strings.Replace(o.State, "_", " ", -1))})

	if o.SupportEmail != "" {
		t.Append([]string{"Email", o.SupportEmail})
	}

	if o.Services != nil {
		buf := bytes.NewBuffer(nil)
		st := table.NewEmbeddedTable(buf)
		st.SetHeader([]string{" "})
		for name, state := range o.Services {
			st.Append([]string{name, strings.ToUpper(strings.Replace(state, "_", " ", -1))})
		}
		st.Render()
		t.Append([]string{"Services", buf.String()})
	}

	if o.Incidents != nil {
		buf := bytes.NewBuffer(nil)
		it := table.NewEmbeddedTable(buf)
		for _, i := range o.Incidents {
			it.Append([]string{i.Title,
				strings.ToUpper(strings.Replace(i.State, "_", " ", -1)),
				i.Status,
				formatSchedule(i.StartDate, i.EndDate),
			})
		}
		it.Render()
		t.Append([]string{"Incidents", buf.String()})
	}

	if o.Maintenances != nil {
		buf := bytes.NewBuffer(nil)
		mt := table.NewEmbeddedTable(buf)
		for _, m := range o.Maintenances {
			mt.Append([]string{m.Title, m.Status, formatSchedule(m.StartDate, m.EndDate)})
		}
		mt.Render()
		t.Append([]string{"Maintenances", buf.String()})
	}

	t.Render()
}

func init() {
	runstatusCmd.AddCommand(&cobra.Command{
		Use:   "show <page name>",
		Short: "Show a runstat.us page details",
		Long: fmt.Sprintf(`This command shows a runstat.us page details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&runstatusPageShowOutput{}), ", ")),
		Aliases: gShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("show expects a page name")
			}

			return output(showRunstatusPage(args[0]))
		},
	})
}

func showRunstatusPage(name string) (outputter, error) {
	page, err := csRunstatus.GetRunstatusPage(gContext, egoscale.RunstatusPage{Subdomain: name})
	if err != nil {
		return nil, err
	}

	out := runstatusPageShowOutput{
		ID:           page.ID,
		Name:         page.Subdomain,
		URL:          page.PublicURL,
		Timezone:     page.TimeZone,
		State:        page.State,
		Title:        page.Title,
		CustomDomain: page.Domain,
		SupportEmail: page.SupportEmail,
	}

	if len(page.Services) > 0 {
		out.Services = make(map[string]string)
		for _, service := range page.Services {
			out.Services[service.Name] = service.State
		}
	}

	if n := len(page.Incidents); n > 0 {
		out.Incidents = make([]runstatusPageIncidentShowOutput, n)
		for i, incident := range page.Incidents {
			out.Incidents[i] = runstatusPageIncidentShowOutput{
				Title:     incident.Title,
				State:     incident.State,
				Status:    incident.Status,
				StartDate: incident.StartDate,
				EndDate:   incident.EndDate,
			}
		}
	}

	if n := len(page.Maintenances); n > 0 {
		out.Maintenances = make([]runstatusPageMaintenanceShowOutput, n)
		for i, maintenance := range page.Maintenances {
			out.Maintenances[i] = runstatusPageMaintenanceShowOutput{
				Title:     maintenance.Title,
				Status:    maintenance.Status,
				StartDate: maintenance.StartDate,
				EndDate:   maintenance.EndDate,
			}
		}
	}

	return &out, nil
}
