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

type dbServiceClickhouseComponentShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	SSL       *bool  `json:"ssl,omitempty"`
	Usage     string `json:"usage"`
}

type dbServiceClickhouseUserShowOutput struct {
	Username string `json:"username,omitempty"`
	UUID     string `json:"uuid,omitempty"`
}

type dbServiceClickhouseShowOutput struct {
	Components     []dbServiceClickhouseComponentShowOutput     `json:"components"`
	ConnectionInfo *dbServiceClickhouseConnectionInfoShowOutput `json:"connection_info,omitempty"`
	IPFilter       []string                                     `json:"ip_filter"`
	PrometheusURI  *dbServiceClickhousePrometheusURIShowOutput  `json:"prometheus_uri,omitempty"`
	URI            string                                       `json:"uri"`
	URIParams      map[string]interface{}                       `json:"uri_params"`
	Users          []dbServiceClickhouseUserShowOutput          `json:"users"`
	Version        string                                       `json:"version"`
}

type dbServiceClickhouseConnectionInfoShowOutput struct {
	URI            []string `json:"uri,omitempty"`
	MysqlURI       string   `json:"mysql_uri,omitempty"`
	ArrowflightURI string   `json:"arrowflight_uri,omitempty"`
}

type dbServiceClickhousePrometheusURIShowOutput struct {
	Host string `json:"host"`
	Port int64  `json:"port"`
}

func formatDatabaseServiceClickhouseTable(t *table.Table, o *dbServiceClickhouseShowOutput) {
	t.Append([]string{"URI", o.URI})
	t.Append([]string{"IP Filter", strings.Join(o.IPFilter, ", ")})
	t.Append([]string{"Version", o.Version})

	if o.ConnectionInfo != nil && len(o.ConnectionInfo.URI) > 0 {
		t.Append([]string{"Connection URIs", strings.Join(o.ConnectionInfo.URI, ", ")})
	}
	if o.ConnectionInfo != nil && o.ConnectionInfo.MysqlURI != "" {
		t.Append([]string{"MySQL URI", o.ConnectionInfo.MysqlURI})
	}
	if o.ConnectionInfo != nil && o.ConnectionInfo.ArrowflightURI != "" {
		t.Append([]string{"ArrowFlight URI", o.ConnectionInfo.ArrowflightURI})
	}

	t.Append([]string{"Components", func() string {
		buf := bytes.NewBuffer(nil)
		ct := table.NewEmbeddedTable(buf)
		ct.SetHeader([]string{" ", "Address", "Route", "Usage"})
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
						users[i] = o.Users[i].Username
					}
					return users
				}(),
				"\n")
		}
		return "n/a"
	}()})
}

func (c *dbaasServiceShowCmd) showDatabaseServiceClickhouse(ctx context.Context) (output.Outputter, error) {

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return nil, err
	}

	databaseService, err := client.GetDBAASServiceClickhouse(ctx, c.Name)
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
		case "clickhouse":
			out, err := json.MarshalIndent(databaseService.ClickhouseSettings, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON: %w", err)
			}
			fmt.Println(string(out))
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(clickhouseSettings, ", "),
			)
		}
		return nil, nil

	case c.ShowURI:
		// Assemble the connection URI from uri_params plus the revealed
		// avnadmin password (ConnectionInfo.URI is not populated for
		// ClickHouse; follow the pg pattern).
		uriParams := databaseService.URIParams

		creds, err := client.RevealDBAASClickhouseUserPassword(
			ctx,
			string(databaseService.Name),
			uriParams["user"].(string),
		)
		if err != nil {
			return nil, err
		}

		uri := fmt.Sprintf(
			"clickhouse://%s:%s@%s:%v/%v",
			uriParams["user"],
			creds.Password,
			uriParams["host"],
			uriParams["port"],
			uriParams["dbname"],
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
		TerminationProtection: utils.DefaultBool(databaseService.TerminationProtection, false),

		Maintenance: func() (v *dbServiceMaintenanceShowOutput) {
			if databaseService.Maintenance != nil {
				v = &dbServiceMaintenanceShowOutput{
					DOW:  string(databaseService.Maintenance.Dow),
					Time: databaseService.Maintenance.Time,
				}
			}
			return
		}(),

		Clickhouse: &dbServiceClickhouseShowOutput{
			Components: func() (v []dbServiceClickhouseComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range databaseService.Components {
						v = append(v, dbServiceClickhouseComponentShowOutput{
							Component: c.Component,
							Host:      c.Host,
							Port:      c.Port,
							Route:     string(c.Route),
							SSL:       c.SSL,
							Usage:     string(c.Usage),
						})
					}
				}
				return
			}(),

			ConnectionInfo: func() (v *dbServiceClickhouseConnectionInfoShowOutput) {
				if databaseService.ConnectionInfo != nil {
					v = &dbServiceClickhouseConnectionInfoShowOutput{
						URI:            databaseService.ConnectionInfo.URI,
						MysqlURI:       databaseService.ConnectionInfo.MysqlURI,
						ArrowflightURI: databaseService.ConnectionInfo.ArrowflightURI,
					}
				}
				return
			}(),

			IPFilter: func() (v []string) {
				if databaseService.IPFilter != nil {
					v = databaseService.IPFilter
				}
				return
			}(),

			PrometheusURI: func() (v *dbServiceClickhousePrometheusURIShowOutput) {
				if databaseService.PrometheusURI != nil {
					v = &dbServiceClickhousePrometheusURIShowOutput{
						Host: databaseService.PrometheusURI.Host,
						Port: databaseService.PrometheusURI.Port,
					}
				}
				return
			}(),

			URI:       databaseService.URI,
			URIParams: databaseService.URIParams,

			Users: func() (v []dbServiceClickhouseUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range databaseService.Users {
						v = append(v, dbServiceClickhouseUserShowOutput{
							Username: string(u.Username),
							UUID:     string(u.Uuid),
						})
					}
				}
				return
			}(),

			Version: databaseService.Version,
		},
	}

	return &out, nil
}
