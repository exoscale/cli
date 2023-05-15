package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mitchellh/go-wordwrap"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
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

func (c *dbaasServiceShowCmd) showDatabaseServiceGrafana(ctx context.Context) (output.Outputter, error) {
	res, err := globalstate.EgoscaleClient.GetDbaasServiceGrafanaWithResponse(ctx, oapi.DbaasServiceName(c.Name))
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return nil, fmt.Errorf("resource not found in zone %q", c.Zone)
		}
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
		case "grafana":
			serviceSettings = databaseService.GrafanaSettings
		default:
			return nil, fmt.Errorf(
				"invalid settings value %q, expected one of: %s",
				c.ShowSettings,
				strings.Join(grafanaSettings, ", "),
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
		fmt.Println(utils.DefaultString(databaseService.Uri, ""))
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

		Grafana: &dbServiceGrafanaShowOutput{
			Components: func() (v []dbServiceGrafanaComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range *databaseService.Components {
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
				if databaseService.IpFilter != nil {
					v = *databaseService.IpFilter
				}
				return
			}(),

			URI:       *databaseService.Uri,
			URIParams: *databaseService.UriParams,

			Users: func() (v []dbServiceGrafanaUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range *databaseService.Users {
						v = append(v, dbServiceGrafanaUserShowOutput{
							Password: utils.DefaultString(u.Password, ""),
							Type:     utils.DefaultString(u.Type, ""),
							Username: utils.DefaultString(u.Username, ""),
						})
					}
				}
				return
			}(),

			Version: utils.DefaultString(databaseService.Version, ""),
		},
	}

	return &out, nil
}
