package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mitchellh/go-wordwrap"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbServiceValkeyComponentShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	Usage     string `json:"usage"`
}

type dbServiceValkeyUserShowOutput struct {
	Password string `json:"password,omitempty"`
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
}

type dbServiceValkeyShowOutput struct {
	Components []dbServiceValkeyComponentShowOutput `json:"components"`
	IPFilter   []string                             `json:"ip_filter"`
	URI        string                               `json:"uri"`
	URIParams  map[string]interface{}               `json:"uri_params"`
	Users      []dbServiceValkeyUserShowOutput      `json:"users"`
}

func formatDatabaseServiceValkeyTable(t *table.Table, o *dbServiceValkeyShowOutput) {
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

func (c *dbaasServiceShowCmd) showDatabaseServiceValkey(ctx context.Context) (output.Outputter, error) {

	client, err := switchClientZoneV3(
		ctx,
		globalstate.EgoscaleV3Client,
		v3.ZoneName(c.Zone),
	)
	databaseService, err := client.GetDBAASServiceValkey(ctx, c.Name)
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
		var serviceSettings *v3.JSONSchemaValkey

		switch c.ShowSettings {
		case "valkey":
			serviceSettings = databaseService.ValkeySettings
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(valkeySettings, ", "),
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
		uriParams := databaseService.URIParams

		creds, err := client.RevealDBAASValkeyUserPassword(
			ctx,
			string(databaseService.Name),
			uriParams["user"].(string),
		)
		if err != nil {
			return nil, err
		}

		// Build URI
		uri := fmt.Sprintf(
			"valkeys://%s:%s@%s:%s",
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

		Valkey: &dbServiceValkeyShowOutput{
			Components: func() (v []dbServiceValkeyComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range databaseService.Components {
						v = append(v, dbServiceValkeyComponentShowOutput{
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
				if databaseService.IPFilter != nil {
					v = databaseService.IPFilter
				}
				return
			}(),

			URI:       databaseService.URI,
			URIParams: databaseService.URIParams,

			Users: func() (v []dbServiceValkeyUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range databaseService.Users {
						v = append(v, dbServiceValkeyUserShowOutput{
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
