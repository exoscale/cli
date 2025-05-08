package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbServiceGrafanaComponentShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	Usage     string `json:"usage"`
}

type dbServiceGrafanaUserShowOutput struct {
	Password string `json:"password,omitempty"`
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
}

type dbServiceGrafanaShowOutput struct {
	Components []dbServiceGrafanaComponentShowOutput `json:"components"`
	IPFilter   []string                              `json:"ip_filter"`
	URI        string                                `json:"uri"`
	URIParams  map[string]interface{}                `json:"uri_params"`
	Users      []dbServiceGrafanaUserShowOutput      `json:"users"`
	Version    string                                `json:"version"`
}

func formatDatabaseServiceGrafanaTable(t *table.Table, o *dbServiceGrafanaShowOutput) {
	t.Append([]string{"Version", o.Version})
	t.Append([]string{"URI", redactDatabaseServiceURI(o.URI)})
	t.Append([]string{"IP Filter", strings.Join(o.IPFilter, ", ")})

	t.Append([]string{"Components", func() string {
		buf := bytes.NewBuffer(nil)
		ct := table.NewEmbeddedTable(buf)
		ct.SetHeader([]string{" "})
		for _, c := range o.Components {
			ct.Append([]string{
				c.Component,
				fmt.Sprintf("%s:%d", c.Host, c.Port),
				"route:" + c.Route,
				"usage:" + c.Usage,
			})
		}
		ct.Render()

		return buf.String()
	}()})

	t.Append([]string{"Users", func() string {
		if len(o.Users) > 0 {
			return strings.Join(
				func() []string {
					users := make([]string, len(o.Users))
					for i := range o.Users {
						users[i] = fmt.Sprintf("%s (%s)", o.Users[i].Username, o.Users[i].Type)
					}
					return users
				}(),
				"\n")
		}
		return "n/a"
	}()})
}

func (c *dbaasServiceShowCmd) showDatabaseServiceGrafana(ctx context.Context) (output.Outputter, error) {

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return nil, err
	}

	res, err := client.GetDBAASServiceGrafana(ctx, c.Name)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return nil, fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return nil, err
	}
	svc := *res

	switch {
	case c.ShowBackups:
		out := make(dbServiceBackupListOutput, 0)
		if svc.Backups != nil {
			for _, b := range svc.Backups {
				out = append(out, dbServiceBackupListItemOutput{
					Date: b.BackupTime,
					Name: b.BackupName,
					Size: b.DataSize,
				})
			}
		}
		return &out, nil

	case c.ShowNotifications:
		out := make(dbServiceNotificationListOutput, 0)
		if svc.Notifications != nil {
			for _, n := range svc.Notifications {
				out = append(out, dbServiceNotificationListItemOutput{
					Level:   string(n.Level),
					Message: wordwrap.WrapString(n.Message, 50),
					Type:    string(n.Type),
				})
			}
		}
		return &out, nil

	case c.ShowSettings != "":

		switch c.ShowSettings {
		case "grafana":
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(grafanaSettings, ", "),
			)
		}

		out, err := json.MarshalIndent(svc.GrafanaSettings, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Println(string(out))

		return nil, nil

	case c.ShowURI:
		fmt.Println(utils.DefaultString(&svc.URI, ""))
		return nil, nil
	}

	out := dbServiceShowOutput{
		Zone:                  c.Zone,
		Name:                  string(svc.Name),
		Type:                  string(svc.Type),
		Plan:                  svc.Plan,
		CreationDate:          svc.CreatedAT,
		Nodes:                 svc.NodeCount,
		NodeCPUs:              svc.NodeCPUCount,
		NodeMemory:            svc.NodeMemory,
		UpdateDate:            svc.UpdatedAT,
		DiskSize:              svc.DiskSize,
		State:                 string(svc.State),
		TerminationProtection: *svc.TerminationProtection,

		Maintenance: func() (v *dbServiceMaintenanceShowOutput) {
			if svc.Maintenance != nil {
				v = &dbServiceMaintenanceShowOutput{
					DOW:  string(svc.Maintenance.Dow),
					Time: svc.Maintenance.Time,
				}
			}
			return
		}(),

		Grafana: &dbServiceGrafanaShowOutput{
			Components: func() (v []dbServiceGrafanaComponentShowOutput) {
				if svc.Components != nil {
					for _, c := range svc.Components {
						v = append(v, dbServiceGrafanaComponentShowOutput{
							Component: c.Component,
							Host:      c.Host,
							Port:      c.Port,
							Route:     string(c.Route),
							Usage:     string(c.Usage),
						})
					}
				}
				return
			}(),

			IPFilter: func() (v []string) {
				if svc.IPFilter != nil {
					v = svc.IPFilter
				}
				return
			}(),

			URI:       svc.URI,
			URIParams: svc.URIParams,

			Users: func() (v []dbServiceGrafanaUserShowOutput) {
				if svc.Users != nil {
					for _, u := range svc.Users {
						v = append(v, dbServiceGrafanaUserShowOutput{
							Password: u.Password,
							Type:     u.Type,
							Username: u.Username,
						})
					}
				}
				return
			}(),

			Version: svc.Version,
		},
	}

	return &out, nil
}
