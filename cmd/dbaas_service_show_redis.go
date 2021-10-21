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

type dbServiceRedisComponentShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	Usage     string `json:"usage"`
}

type dbServiceRedisUserShowOutput struct {
	Password string `json:"password,omitempty"`
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
}

type dbServiceRedisShowOutput struct {
	Components []dbServiceRedisComponentShowOutput `json:"components"`
	IPFilter   []string                            `json:"ip_filter"`
	URI        string                              `json:"uri"`
	URIParams  map[string]interface{}              `json:"uri_params"`
	Users      []dbServiceRedisUserShowOutput      `json:"users"`
	Version    string                              `json:"version"`
}

func formatDatabaseServiceRedisTable(t *table.Table, o *dbServiceRedisShowOutput) {
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

func (c *dbServiceShowCmd) showDatabaseServiceRedis(ctx context.Context) (outputter, error) {
	res, err := cs.GetDbaasServiceRedisWithResponse(ctx, oapi.DbaasServiceName(c.Name))
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
		case "redis":
			serviceSettings = databaseService.RedisSettings
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(redisSettings, ", "),
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

		Redis: &dbServiceRedisShowOutput{
			Components: func() (v []dbServiceRedisComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range *databaseService.Components {
						v = append(v, dbServiceRedisComponentShowOutput{
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
				if databaseService.IpFilter != nil {
					v = *databaseService.IpFilter
				}
				return
			}(),

			URI:       *databaseService.Uri,
			URIParams: *databaseService.UriParams,

			Users: func() (v []dbServiceRedisUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range *databaseService.Users {
						v = append(v, dbServiceRedisUserShowOutput{
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
