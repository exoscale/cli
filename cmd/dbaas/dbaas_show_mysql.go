package dbaas

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbServiceMysqlComponentShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	Usage     string `json:"usage"`
}

type dbServiceMysqlUserShowOutput struct {
	Password string `json:"password,omitempty"`
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
}

type dbServiceMysqlShowOutput struct {
	BackupSchedule string                              `json:"backup_schedule"`
	Components     []dbServiceMysqlComponentShowOutput `json:"components"`
	Databases      []string                            `json:"databases"`
	IPFilter       []string                            `json:"ip_filter"`
	URI            string                              `json:"uri"`
	URIParams      map[string]interface{}              `json:"uri_params"`
	Users          []dbServiceMysqlUserShowOutput      `json:"users"`
	Version        string                              `json:"version"`
}

func formatDatabaseServiceMysqlTable(t *table.Table, o *dbServiceMysqlShowOutput) {
	t.Append([]string{"Version", o.Version})
	t.Append([]string{"Backup Schedule", o.BackupSchedule})
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

	t.Append([]string{"Databases", func() string {
		if len(o.Databases) > 0 {
			return strings.Join(
				func() []string {
					dbs := make([]string, len(o.Databases))
					copy(dbs, o.Databases)
					return dbs
				}(),
				"\n")
		}
		return "n/a"
	}()})
}

func (c *dbaasServiceShowCmd) showDatabaseServiceMysql(ctx context.Context) (output.Outputter, error) {

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return nil, err
	}

	databaseService, err := client.GetDBAASServiceMysql(ctx, c.Name)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return nil, fmt.Errorf("resource not found in zone %q", c.Zone)
		}
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
		case "mysql":
			out, err := json.MarshalIndent(databaseService.MysqlSettings, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON: %w", err)
			}
			fmt.Println(string(out))
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(mysqlSettings, ", "),
			)
		}

		return nil, nil

	case c.ShowURI:
		// Read password from dedicated endpoint
		client, err := exocmd.SwitchClientZoneV3(
			ctx,
			globalstate.EgoscaleV3Client,
			v3.ZoneName(c.Zone),
		)
		if err != nil {
			return nil, err
		}

		uriParams := databaseService.URIParams

		creds, err := client.RevealDBAASMysqlUserPassword(
			ctx,
			string(databaseService.Name),
			uriParams["user"].(string),
		)
		if err != nil {
			return nil, err
		}

		// Build URI
		uri := fmt.Sprintf(
			"mysql://%s:%s@%s:%s/%s?ssl-mode=%s",
			uriParams["user"],
			creds.Password,
			uriParams["host"],
			uriParams["port"],
			uriParams["dbname"],
			uriParams["ssl-mode"],
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

		Mysql: &dbServiceMysqlShowOutput{
			BackupSchedule: func() (v string) {
				if databaseService.BackupSchedule != nil {
					v = fmt.Sprintf(
						"%02d:%02d",
						databaseService.BackupSchedule.BackupHour,
						databaseService.BackupSchedule.BackupMinute,
					)
				}
				return
			}(),

			Components: func() (v []dbServiceMysqlComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range databaseService.Components {
						v = append(v, dbServiceMysqlComponentShowOutput{
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

			Databases: func() (v []string) {
				v = make([]string, len(databaseService.Databases))
				for i, d := range databaseService.Databases {
					v[i] = string(d)
				}

				return
			}(),

			IPFilter: func() (v []string) {
				if databaseService.IPFilter != nil {
					v = databaseService.IPFilter
				}
				return
			}(),

			URI:       databaseService.URI,
			URIParams: databaseService.URIParams,

			Users: func() (v []dbServiceMysqlUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range databaseService.Users {
						v = append(v, dbServiceMysqlUserShowOutput{
							Password: u.Password,
							Type:     u.Type,
							Username: u.Username,
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
