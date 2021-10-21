package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/mitchellh/go-wordwrap"
)

type dbServicePGConnectionPool struct {
	ConnectionURI string `json:"connection_uri"`
	Database      string `json:"database"`
	Mode          string `json:"mode"`
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	Username      string `json:"username"`
}

type dbServicePGComponentShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	Usage     string `json:"usage"`
}

type dbServicePGUserShowOutput struct {
	AllowReplication bool   `json:"allow_replication,omitempty"`
	Password         string `json:"password,omitempty"`
	Type             string `json:"type,omitempty"`
	Username         string `json:"username,omitempty"`
}

type dbServicePGShowOutput struct {
	BackupSchedule  string                           `json:"backup_schedule"`
	Components      []dbServicePGComponentShowOutput `json:"components"`
	ConnectionPools []dbServicePGConnectionPool      `json:"connection_pools"`
	IPFilter        []string                         `json:"ip_filter"`
	URI             string                           `json:"uri"`
	URIParams       map[string]interface{}           `json:"uri_params"`
	Users           []dbServicePGUserShowOutput      `json:"users"`
	Version         string                           `json:"version"`
}

func formatDatabaseServicePGTable(t *table.Table, o *dbServicePGShowOutput) {
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

	if len(o.ConnectionPools) > 0 {
		t.Append([]string{"Connection Pools", func() string {
			buf := bytes.NewBuffer(nil)
			pt := table.NewEmbeddedTable(buf)
			pt.SetHeader([]string{" "})
			for _, pool := range o.ConnectionPools {
				pt.Append([]string{
					pool.Name,
					"database:" + pool.Database,
					"size:" + fmt.Sprint(pool.Size),
					"mode:" + pool.Mode,
				})
			}
			pt.Render()

			return buf.String()
		}()})
	}

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

func (c *dbServiceShowCmd) showDatabaseServicePG(ctx context.Context) (outputter, error) {
	res, err := cs.GetDbaasServicePgWithResponse(ctx, oapi.DbaasServiceName(c.Name))
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API request error: unexpected status %s", res.Status())
	}
	databaseService := res.JSON200

	switch {
	case c.ShowBackups:
		out := make(dbServiceBackupListOutput, 0)
		if databaseService.Backups != nil {
			for _, b := range *databaseService.Backups {
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
			for _, n := range *databaseService.Notifications {
				out = append(out, dbServiceNotificationListItemOutput{
					Level:   string(n.Level),
					Message: wordwrap.WrapString(n.Message, 50),
					Type:    string(n.Type),
				})
			}
		}
		return &out, nil

	case c.ShowSettings != "":
		var serviceSettings *map[string]interface{}

		switch c.ShowSettings {
		case "pg":
			serviceSettings = databaseService.PgSettings
		case "pgbouncer":
			serviceSettings = databaseService.PgbouncerSettings
		case "pglookout":
			serviceSettings = databaseService.PglookoutSettings
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(kafkaSettings, ", "),
			)
		}

		if serviceSettings != nil {
			out, err := json.MarshalIndent(serviceSettings, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON: %w", err)
			}
			fmt.Println(string(out))
		}

		return nil, nil

	case c.ShowURI:
		fmt.Println(defaultString(databaseService.Uri, ""))
		return nil, nil
	}

	out := dbServiceShowOutput{
		Zone:                  c.Zone,
		Name:                  string(databaseService.Name),
		Type:                  string(databaseService.Type),
		Plan:                  databaseService.Plan,
		CreationDate:          *databaseService.CreatedAt,
		Nodes:                 *databaseService.NodeCount,
		NodeCPUs:              *databaseService.NodeCpuCount,
		NodeMemory:            *databaseService.NodeMemory,
		UpdateDate:            *databaseService.UpdatedAt,
		DiskSize:              *databaseService.DiskSize,
		State:                 string(*databaseService.State),
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

		PG: &dbServicePGShowOutput{
			BackupSchedule: func() (v string) {
				if databaseService.BackupSchedule != nil {
					v = fmt.Sprintf(
						"%02d:%02d",
						*databaseService.BackupSchedule.BackupHour,
						*databaseService.BackupSchedule.BackupMinute,
					)
				}
				return
			}(),

			Components: func() (v []dbServicePGComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range *databaseService.Components {
						v = append(v, dbServicePGComponentShowOutput{
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

			ConnectionPools: func() (v []dbServicePGConnectionPool) {
				if databaseService.ConnectionPools != nil {
					for _, pool := range *databaseService.ConnectionPools {
						v = append(v, dbServicePGConnectionPool{
							ConnectionURI: pool.ConnectionUri,
							Database:      pool.Database,
							Mode:          string(pool.Mode),
							Name:          pool.Name,
							Size:          pool.Size,
							Username:      pool.Username,
						})
					}
				}
				return
			}(),

			IPFilter: func() (v []string) {
				if databaseService.IpFilter != nil {
					v = *databaseService.IpFilter
				}
				return
			}(),

			URI:       *databaseService.Uri,
			URIParams: *databaseService.UriParams,

			Users: func() (v []dbServicePGUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range *databaseService.Users {
						v = append(v, dbServicePGUserShowOutput{
							AllowReplication: func() (v bool) {
								if u.AccessControl != nil {
									v = defaultBool(u.AccessControl.AllowReplication, false)
								}
								return
							}(),
							Password: defaultString(u.Password, ""),
							Type:     defaultString(u.Type, ""),
							Username: defaultString(u.Username, ""),
						})
					}
				}
				return
			}(),
		},
	}

	return &out, nil
}
