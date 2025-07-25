package dbaas

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-wordwrap"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbServiceOpensearchComponentsShowOutput struct {
	Component string `json:"component"`
	Host      string `json:"host"`
	Port      int64  `json:"port"`
	Route     string `json:"route"`
	Usage     string `json:"usage"`
}

type dbServiceOpensearchConnectionInfoShowOutput struct {
	DashboardURI string   `json:"dashboard-uri,omitempty"`
	Password     string   `json:"password,omitempty"`
	URI          []string `json:"uri,omitempty"`
	Username     string   `json:"username,omitempty"`
}

type dbServiceOpensearchIndexTemplateShowOutput struct {
	MappingNestedObjectsLimit int64 `json:"mapping-nested-objects-limit,omitempty"`
	NumberOfReplicas          int64 `json:"number-of-replicas,omitempty"`
	NumberOfShards            int64 `json:"number-of-shards,omitempty"`
}

type dbServiceOpensearchDashboardShowOutput struct {
	Enabled                  bool  `json:"enabled,omitempty"`
	MaxOldSpaceSize          int64 `json:"max-old-space-size,omitempty"`
	OpensearchRequestTimeout int64 `json:"opensearch-request-timeout,omitempty"`
}

type dbServiceOpensearchUserShowOutput struct {
	Password string `json:"password,omitempty"`
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
}

type dbServiceOpensearchIndexPatternShowOutput struct {
	MaxIndexCount    int64  `json:"max-index-count,omitempty"`
	Pattern          string `json:"pattern,omitempty"`
	SortingAlgorithm string `json:"sorting-algorithm,omitempty"`
}

type dbServiceOpensearchShowOutput struct {
	IPFilter                 []string                                    `json:"ip_filter"`
	URI                      string                                      `json:"uri"`
	URIParams                map[string]interface{}                      `json:"uri_params"`
	Version                  string                                      `json:"version"`
	Components               []dbServiceOpensearchComponentsShowOutput   `json:"components,omitempty"`
	ConnectionInfo           dbServiceOpensearchConnectionInfoShowOutput `json:"connection-info,omitempty"`
	Description              string                                      `json:"description,omitempty"`
	IndexPatterns            []dbServiceOpensearchIndexPatternShowOutput `json:"index-patterns,omitempty"`
	IndexTemplate            *dbServiceOpensearchIndexTemplateShowOutput `json:"index-template,omitempty"`
	KeepIndexRefreshInterval bool                                        `json:"keep-index-refresh-interval,omitempty"`
	Dashboard                *dbServiceOpensearchDashboardShowOutput     `json:"opensearch-dashboards,omitempty"`
	Users                    []dbServiceOpensearchUserShowOutput         `json:"users,omitempty"`
}

func formatDatabaseServiceOpensearchTable(t *table.Table, o *dbServiceOpensearchShowOutput) {
	t.Append([]string{"Version", o.Version})
	t.Append([]string{"URI", redactDatabaseServiceURI(o.URI)})
	t.Append([]string{"IP Filter", strings.Join(o.IPFilter, ", ")})

	buf := bytes.NewBuffer(nil)
	et := table.NewEmbeddedTable(buf)
	for _, c := range o.Components {
		et.SetHeader([]string{"Name", "host:port", "route", "usage"})
		et.Append([]string{c.Component, fmt.Sprintf("%s:%d", c.Host, c.Port), c.Route, c.Usage})
	}
	et.Render()
	t.Append([]string{"Components", buf.String()})

	t.Append([]string{"Description", o.Description})

	buf.Reset()
	et = table.NewEmbeddedTable(buf)
	if o.IndexPatterns != nil {
		et.SetHeader([]string{"Pattern", "Max Index Count", "Sorting Algorithm"})
		for _, i := range o.IndexPatterns {
			et.Append([]string{i.Pattern, strconv.FormatInt(i.MaxIndexCount, 10), i.SortingAlgorithm})
		}
	}
	et.Render()
	t.Append([]string{"IndexPatterns", buf.String()})

	var indexTemplate string
	if o.IndexTemplate != nil {
		indexTemplate = fmt.Sprintf("MappingNestedObjectsLimit:%d NumberOfReplicas:%d NumberOfShards:%d",
			o.IndexTemplate.MappingNestedObjectsLimit,
			o.IndexTemplate.NumberOfReplicas,
			o.IndexTemplate.NumberOfShards)
	}
	t.Append([]string{"IndexTemplate", indexTemplate})

	t.Append([]string{"KeepIndexRefreshInterval", fmt.Sprint(o.KeepIndexRefreshInterval)})

	var dashboard string
	if o.Dashboard != nil {
		dashboard = fmt.Sprintf("Enabled:%v MaxOldSpaceSize:%d OpensearchRequestTimeout:%d",
			o.Dashboard.Enabled, o.Dashboard.MaxOldSpaceSize, o.Dashboard.OpensearchRequestTimeout)
	}
	t.Append([]string{"Dashboard", dashboard})

	var users string
	for _, u := range o.Users {
		users += fmt.Sprintf("%s (%s)\n", u.Username, u.Type)
	}
	t.Append([]string{"Users", users})
}

func (c *dbaasServiceShowCmd) showDatabaseServiceOpensearch(ctx context.Context) (output.Outputter, error) {

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return nil, err
	}

	res, err := client.GetDBAASServiceOpensearch(ctx, c.Name)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return nil, fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return nil, err
	}

	switch {
	case c.ShowBackups:
		return opensearchShowBackups(res)
	case c.ShowNotifications:
		return opensearchShowNotifications(res)
	case c.ShowSettings != "":
		return nil, opensearchShowSettings(c.ShowSettings, res)
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

		uriParams := res.URIParams

		creds, err := client.RevealDBAASOpensearchUserPassword(
			ctx,
			string(res.Name),
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
	default:
		return opensearchShowDatabase(res, c.Zone)
	}
}

func opensearchShowSettings(setting string, db *v3.DBAASServiceOpensearch) error {

	switch setting {
	case "opensearch":
		out, err := json.MarshalIndent(db.OpensearchSettings, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Println(string(out))

	default:
		return fmt.Errorf("invalid settings value %q, expected one of: %s", setting, strings.Join(opensearchSettings, ", "))
	}

	return nil
}

func opensearchShowNotifications(db *v3.DBAASServiceOpensearch) (output.Outputter, error) {
	out := make(dbServiceNotificationListOutput, 0)
	if db.Notifications != nil {
		for _, n := range db.Notifications {
			out = append(out, dbServiceNotificationListItemOutput{
				Level:   string(n.Level),
				Message: wordwrap.WrapString(n.Message, 50),
				Type:    string(n.Type),
			})
		}
	}
	return &out, nil
}

func opensearchShowBackups(db *v3.DBAASServiceOpensearch) (output.Outputter, error) {
	if db.Backups == nil {
		return &dbServiceBackupListOutput{}, nil
	}

	out := make(dbServiceBackupListOutput, 0, len(db.Backups))
	for _, b := range db.Backups {
		out = append(out, dbServiceBackupListItemOutput{
			Date: b.BackupTime,
			Name: b.BackupName,
			Size: b.DataSize,
		})
	}

	return &out, nil
}

func opensearchShowDatabase(db *v3.DBAASServiceOpensearch, zone string) (output.Outputter, error) {
	var components []dbServiceOpensearchComponentsShowOutput
	if db.Components != nil {
		for _, c := range db.Components {
			components = append(components, dbServiceOpensearchComponentsShowOutput{
				Component: c.Component,
				Host:      c.Host,
				Port:      c.Port,
				Route:     string(c.Route),
				Usage:     string(c.Usage),
			})
		}
	}

	var indexPatterns []dbServiceOpensearchIndexPatternShowOutput
	if db.IndexPatterns != nil {
		for _, i := range db.IndexPatterns {
			indexPatterns = append(indexPatterns, dbServiceOpensearchIndexPatternShowOutput{
				MaxIndexCount:    i.MaxIndexCount,
				Pattern:          i.Pattern,
				SortingAlgorithm: string(i.SortingAlgorithm),
			})
		}
	}

	var indexTemplate *dbServiceOpensearchIndexTemplateShowOutput
	if db.IndexTemplate != nil {
		indexTemplate = &dbServiceOpensearchIndexTemplateShowOutput{
			MappingNestedObjectsLimit: db.IndexTemplate.MappingNestedObjectsLimit,
			NumberOfReplicas:          db.IndexTemplate.NumberOfReplicas,
			NumberOfShards:            db.IndexTemplate.NumberOfShards,
		}
	}

	var dashboard *dbServiceOpensearchDashboardShowOutput
	if db.OpensearchDashboards != nil {
		dashboard = &dbServiceOpensearchDashboardShowOutput{
			Enabled:                  utils.DefaultBool(db.OpensearchDashboards.Enabled, false),
			MaxOldSpaceSize:          db.OpensearchDashboards.MaxOldSpaceSize,
			OpensearchRequestTimeout: db.OpensearchDashboards.OpensearchRequestTimeout,
		}
	}

	var users []dbServiceOpensearchUserShowOutput
	if db.Users != nil {
		for _, u := range db.Users {
			users = append(users, dbServiceOpensearchUserShowOutput{
				Password: u.Password,
				Type:     u.Type,
				Username: u.Username,
			})
		}
	}

	return &dbServiceShowOutput{
		Zone: zone,
		Name: string(db.Name),
		Type: string(db.Type),
		Plan: db.Plan,
		CreationDate: func() time.Time {
			if !db.CreatedAT.IsZero() {
				return db.CreatedAT
			}
			return time.Time{}
		}(),
		Nodes:      db.NodeCount,
		NodeCPUs:   db.NodeCPUCount,
		NodeMemory: db.NodeMemory,
		UpdateDate: func() time.Time {
			if !db.UpdatedAT.IsZero() {
				return db.UpdatedAT
			}
			return time.Time{}
		}(),
		DiskSize: db.DiskSize,
		State: func() string {
			if db.State != "" {
				return string(db.State)
			}
			return ""
		}(),
		TerminationProtection: utils.DefaultBool(db.TerminationProtection, false),

		Maintenance: func() (v *dbServiceMaintenanceShowOutput) {
			if db.Maintenance != nil {
				v = &dbServiceMaintenanceShowOutput{
					DOW:  string(db.Maintenance.Dow),
					Time: db.Maintenance.Time,
				}
			}
			return
		}(),

		Opensearch: &dbServiceOpensearchShowOutput{
			IPFilter: func() (v []string) {
				if db.IPFilter != nil {
					v = db.IPFilter
				}
				return
			}(),
			URI: db.URI,
			URIParams: func() map[string]interface{} {
				if db.URIParams != nil {
					return db.URIParams
				}
				return map[string]interface{}{}
			}(),
			Version:    db.Version,
			Components: components,
			ConnectionInfo: dbServiceOpensearchConnectionInfoShowOutput{
				DashboardURI: db.ConnectionInfo.DashboardURI,
				Password:     db.ConnectionInfo.Password,
				URI: func() []string {
					if db.ConnectionInfo.URI != nil {
						return db.ConnectionInfo.URI
					}
					return []string{}
				}(),
				Username: db.ConnectionInfo.Username,
			},
			Description:              db.Description,
			IndexPatterns:            indexPatterns,
			IndexTemplate:            indexTemplate,
			KeepIndexRefreshInterval: utils.DefaultBool(db.KeepIndexRefreshInterval, false),
			Dashboard:                dashboard,
			Users:                    users,
		},
	}, nil
}
