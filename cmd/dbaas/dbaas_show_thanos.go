package dbaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbServiceThanosComponentShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	Usage     string `json:"usage"`
}

type dbServiceThanosUserShowOutput struct {
	Password string `json:"password,omitempty"`
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
}

type dbServiceThanosConnectionInfoOutput struct {
	QueryFrontendURI       string `json:"query_frontend_uri,omitempty"`
	QueryURI               string `json:"query_uri,omitempty"`
	ReceiverRemoteWriteURI string `json:"receiver_remote_write_uri,omitempty"`
	RulerURI               string `json:"ruler_uri,omitempty"`
}

type dbServiceThanosPrometheusURIOutput struct {
	Host string `json:"host,omitempty"`
	Port int64  `json:"port,omitempty"`
}

type dbServiceThanosShowOutput struct {
	Components     []dbServiceThanosComponentShowOutput `json:"components"`
	ConnectionInfo *dbServiceThanosConnectionInfoOutput `json:"connection_info,omitempty"`
	IPFilter       []string                             `json:"ip_filter"`
	PrometheusURI  *dbServiceThanosPrometheusURIOutput  `json:"prometheus_uri,omitempty"`
	URI            string                               `json:"uri"`
	URIParams      map[string]interface{}               `json:"uri_params"`
	Users          []dbServiceThanosUserShowOutput      `json:"users"`
}

var thanosSettings = []string{"thanos", "compactor", "query", "query-frontend"}

func formatDatabaseServiceThanosTable(t *table.Table, o *dbServiceThanosShowOutput) {
	t.Append([]string{"URI", redactDatabaseServiceURI(o.URI)})
	t.Append([]string{"IP Filter", strings.Join(o.IPFilter, ", ")})

	if o.ConnectionInfo != nil {
		t.Append([]string{"Connection Info", func() string {
			buf := bytes.NewBuffer(nil)
			ct := table.NewEmbeddedTable(buf)
			ct.SetHeader([]string{" "})
			if o.ConnectionInfo.QueryFrontendURI != "" {
				ct.Append([]string{"Query Frontend URI", o.ConnectionInfo.QueryFrontendURI})
			}
			if o.ConnectionInfo.QueryURI != "" {
				ct.Append([]string{"Query URI", o.ConnectionInfo.QueryURI})
			}
			if o.ConnectionInfo.ReceiverRemoteWriteURI != "" {
				ct.Append([]string{"Receiver Remote Write URI", o.ConnectionInfo.ReceiverRemoteWriteURI})
			}
			if o.ConnectionInfo.RulerURI != "" {
				ct.Append([]string{"Ruler URI", o.ConnectionInfo.RulerURI})
			}
			ct.Render()
			return buf.String()
		}()})
	}

	if o.PrometheusURI != nil {
		t.Append([]string{"Prometheus URI", fmt.Sprintf("%s:%d", o.PrometheusURI.Host, o.PrometheusURI.Port)})
	}

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

func (c *dbaasServiceShowCmd) showDatabaseServiceThanos(ctx context.Context) (output.Outputter, error) {

	client, err := exocmd.SwitchClientZoneV3(
		ctx,
		globalstate.EgoscaleV3Client,
		v3.ZoneName(c.Zone),
	)

	if err != nil {
		return nil, err
	}

	databaseService, err := client.GetDBAASServiceThanos(ctx, c.Name)
	if err != nil {
		return nil, err
	}

	switch {
	case c.ShowBackups:
		out := make(dbServiceBackupListOutput, 0)
		if databaseService.Backups != nil {
			for _, b := range databaseService.Backups {
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
		if databaseService.Notifications != nil {
			for _, n := range databaseService.Notifications {
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
		case "thanos":
			out, err := json.MarshalIndent(databaseService.ThanosSettings, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON: %w", err)
			}
			fmt.Println(string(out))
		case "compactor":
			if databaseService.ThanosSettings != nil && databaseService.ThanosSettings.Compactor != nil {
				out, err := json.MarshalIndent(databaseService.ThanosSettings.Compactor, "", "  ")
				if err != nil {
					return nil, fmt.Errorf("unable to marshal JSON: %w", err)
				}
				fmt.Println(string(out))
			}
		case "query":
			if databaseService.ThanosSettings != nil && databaseService.ThanosSettings.Query != nil {
				out, err := json.MarshalIndent(databaseService.ThanosSettings.Query, "", "  ")
				if err != nil {
					return nil, fmt.Errorf("unable to marshal JSON: %w", err)
				}
				fmt.Println(string(out))
			}
		case "query-frontend":
			if databaseService.ThanosSettings != nil && databaseService.ThanosSettings.QueryFrontend != nil {
				out, err := json.MarshalIndent(databaseService.ThanosSettings.QueryFrontend, "", "  ")
				if err != nil {
					return nil, fmt.Errorf("unable to marshal JSON: %w", err)
				}
				fmt.Println(string(out))
			}
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(thanosSettings, ", "),
			)
		}

		return nil, nil

	case c.ShowURI:
		uriParams := databaseService.URIParams

		creds, err := client.RevealDBAASThanosUserPassword(
			ctx,
			string(databaseService.Name),
			uriParams["user"].(string),
		)
		if err != nil {
			return nil, err
		}

		// Build URI
		uri := fmt.Sprintf(
			"https://%s:%s@%s:%s",
			uriParams["user"],
			creds.Password,
			uriParams["host"],
			uriParams["port"],
		)

		fmt.Println(uri)
		return nil, nil
	}

	out := dbServiceShowOutput{
		Zone:                  c.Zone,
		Name:                  string(databaseService.Name),
		Type:                  string(databaseService.Type),
		Plan:                  databaseService.Plan,
		CreationDate:          databaseService.CreatedAT,
		Nodes:                 databaseService.NodeCount,
		NodeCPUs:              databaseService.NodeCPUCount,
		NodeMemory:            databaseService.NodeMemory,
		UpdateDate:            databaseService.UpdatedAT,
		DiskSize:              databaseService.DiskSize,
		State:                 string(databaseService.State),
		TerminationProtection: *databaseService.TerminationProtection,

		Maintenance: func() (v *dbServiceMaintenanceShowOutput) {
			if databaseService.Maintenance != nil {
				v = &dbServiceMaintenanceShowOutput{
					DOW:  string(databaseService.Maintenance.Dow),
					Time: databaseService.Maintenance.Time,
				}
			}
			return
		}(),

		Thanos: &dbServiceThanosShowOutput{
			Components: func() (v []dbServiceThanosComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range databaseService.Components {
						v = append(v, dbServiceThanosComponentShowOutput{
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

			ConnectionInfo: func() *dbServiceThanosConnectionInfoOutput {
				if databaseService.ConnectionInfo != nil {
					return &dbServiceThanosConnectionInfoOutput{
						QueryFrontendURI:       databaseService.ConnectionInfo.QueryFrontendURI,
						QueryURI:               databaseService.ConnectionInfo.QueryURI,
						ReceiverRemoteWriteURI: databaseService.ConnectionInfo.ReceiverRemoteWriteURI,
						RulerURI:               databaseService.ConnectionInfo.RulerURI,
					}
				}
				return nil
			}(),

			IPFilter: func() (v []string) {
				if databaseService.IPFilter != nil {
					v = databaseService.IPFilter
				}
				return
			}(),

			PrometheusURI: func() *dbServiceThanosPrometheusURIOutput {
				if databaseService.PrometheusURI != nil {
					return &dbServiceThanosPrometheusURIOutput{
						Host: databaseService.PrometheusURI.Host,
						Port: databaseService.PrometheusURI.Port,
					}
				}
				return nil
			}(),

			URI:       databaseService.URI,
			URIParams: databaseService.URIParams,

			Users: func() (v []dbServiceThanosUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range databaseService.Users {
						v = append(v, dbServiceThanosUserShowOutput{
							Password: utils.DefaultString(&u.Password, ""),
							Type:     utils.DefaultString(&u.Type, ""),
							Username: utils.DefaultString(&u.Username, ""),
						})
					}
				}
				return
			}(),
		},
	}

	return &out, nil
}
