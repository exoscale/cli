package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/mitchellh/go-wordwrap"
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

type dbServiceOpensearchShowOutput struct {
	IPFilter                 []string                                    `json:"ip_filter"`
	URI                      string                                      `json:"uri"`
	URIParams                map[string]interface{}                      `json:"uri_params"`
	Version                  string                                      `json:"version"`
	Components               []dbServiceOpensearchComponentsShowOutput   `json:"components,omitempty"`
	ConnectionInfo           dbServiceOpensearchConnectionInfoShowOutput `json:"connection-info,omitempty"`
	Description              string                                      `json:"description,omitempty"`
	IndexPatterns            string                                      `json:"index-patterns,omitempty"`
	IndexTemplate            *dbServiceOpensearchIndexTemplateShowOutput `json:"index-template,omitempty"`
	KeepIndexRefreshInterval bool                                        `json:"keep-index-refresh-interval,omitempty"`
	MaxIndexCount            int64                                       `json:"max-index-count,omitempty"`
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
		et.Append([]string{
			c.Component,
			fmt.Sprintf("%s:%d", c.Host, c.Port),
			"route:" + c.Route,
			"usage:" + c.Usage,
		})
	}
	t.Append([]string{"Components", buf.String()})

	t.Append([]string{"Description", o.Description})
	t.Append([]string{"IndexPatterns", o.IndexPatterns})

	buf.Reset()
	et = table.NewEmbeddedTable(buf)
	if o.IndexTemplate != nil {
		et.Append([]string{
			"MappingNestedObjectsLimit:" + strconv.FormatInt(o.IndexTemplate.MappingNestedObjectsLimit, 10),
			"NumberOfReplicas:" + strconv.FormatInt(o.IndexTemplate.NumberOfReplicas, 10),
			"NumberOfShards:" + strconv.FormatInt(o.IndexTemplate.NumberOfShards, 10),
		})
	}
	t.Append([]string{"IndexTemplate", buf.String()})

	t.Append([]string{"KeepIndexRefreshInterval", fmt.Sprint(o.KeepIndexRefreshInterval)})
	t.Append([]string{"MaxIndexCount", strconv.FormatInt(o.MaxIndexCount, 10)})

	buf.Reset()
	et = table.NewEmbeddedTable(buf)
	if o.Dashboard != nil {
		et.Append([]string{
			"Enabled:" + fmt.Sprint(o.Dashboard.Enabled),
			"MaxOldSpaceSize:" + strconv.FormatInt(o.Dashboard.MaxOldSpaceSize, 10),
			"OpensearchRequestTimeout:" + strconv.FormatInt(o.Dashboard.OpensearchRequestTimeout, 10),
		})
	}
	t.Append([]string{"Dashboard", buf.String()})

	var users string
	for _, u := range o.Users {
		users += fmt.Sprintf("%s (%s)", u.Username, u.Type)
	}
	t.Append([]string{"Users", users})

}

func (c *dbaasServiceShowCmd) showDatabaseServiceOpensearch(ctx context.Context) (outputter, error) {
	res, err := cs.GetDbaasServiceOpensearchWithResponse(ctx, oapi.DbaasServiceName(c.Name))
	if err != nil {
		return nil, err
	}
	if res.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API request error: unexpected status %s", res.Status())
	}

	switch {
	case c.ShowBackups:
		return opensearchShowBackups(res.JSON200)
	case c.ShowNotifications:
		return opensearchShowNotifications(res.JSON200)
	case c.ShowSettings != "":
		return nil, opensearchShowSettings(c.ShowSettings, res.JSON200)
	case c.ShowURI:
		fmt.Println(defaultString(res.JSON200.Uri, ""))
		return nil, nil
	default:
		return opensearchShowDatabase(res.JSON200, c.Zone)
	}
}

func opensearchShowSettings(setting string, db *oapi.DbaasServiceOpensearch) error {
	var serviceSettings *map[string]interface{}

	switch setting {
	case "opensearch":
		serviceSettings = db.OpensearchSettings
	default:
		return fmt.Errorf("invalid settings value %q, expected one of: %s", setting, strings.Join(opensearchSettings, ", "))
	}

	if serviceSettings != nil {
		out, err := json.MarshalIndent(serviceSettings, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Println(string(out))
	}

	return nil
}

func opensearchShowNotifications(db *oapi.DbaasServiceOpensearch) (outputter, error) {
	out := make(dbServiceNotificationListOutput, 0)
	if db.Notifications != nil {
		for _, n := range *db.Notifications {
			out = append(out, dbServiceNotificationListItemOutput{
				Level:   string(n.Level),
				Message: wordwrap.WrapString(n.Message, 50),
				Type:    string(n.Type),
			})
		}
	}
	return &out, nil
}

func opensearchShowBackups(db *oapi.DbaasServiceOpensearch) (outputter, error) {
	out := make(dbServiceBackupListOutput, 0)
	if db.Backups != nil {
		for _, b := range *db.Backups {
			out = append(out, dbServiceBackupListItemOutput{
				Date: b.BackupTime,
				Name: b.BackupName,
				Size: b.DataSize,
			})
		}
	}
	return &out, nil
}

func opensearchShowDatabase(db *oapi.DbaasServiceOpensearch, zone string) (outputter, error) {
	var components []dbServiceOpensearchComponentsShowOutput
	if db.Components != nil {
		for _, c := range *db.Components {
			components = append(components, dbServiceOpensearchComponentsShowOutput{
				Component: c.Component,
				Host:      c.Host,
				Port:      c.Port,
				Route:     string(c.Route),
				Usage:     string(c.Usage),
			})
		}
	}

	var indexPatterns []byte
	if db.IndexPatterns != nil {
		var err error
		indexPatterns, err = json.Marshal(*db.IndexPatterns)
		if err != nil {
			return nil, fmt.Errorf("failed to encode IndexPatterns: %s", err)
		}
	}

	var indexTemplate *dbServiceOpensearchIndexTemplateShowOutput
	if db.IndexTemplate != nil {
		indexTemplate = &dbServiceOpensearchIndexTemplateShowOutput{
			MappingNestedObjectsLimit: *db.IndexTemplate.MappingNestedObjectsLimit,
			NumberOfReplicas:          *db.IndexTemplate.NumberOfReplicas,
			NumberOfShards:            *db.IndexTemplate.NumberOfShards,
		}
	}

	var dashboard *dbServiceOpensearchDashboardShowOutput
	if db.OpensearchDashboards != nil {
		dashboard = &dbServiceOpensearchDashboardShowOutput{
			Enabled:                  *db.OpensearchDashboards.Enabled,
			MaxOldSpaceSize:          *db.OpensearchDashboards.MaxOldSpaceSize,
			OpensearchRequestTimeout: *db.OpensearchDashboards.OpensearchRequestTimeout,
		}
	}

	return &dbServiceShowOutput{
		Zone:                  zone,
		Name:                  string(db.Name),
		Type:                  string(db.Type),
		Plan:                  db.Plan,
		CreationDate:          *db.CreatedAt,
		Nodes:                 *db.NodeCount,
		NodeCPUs:              *db.NodeCpuCount,
		NodeMemory:            *db.NodeMemory,
		UpdateDate:            *db.UpdatedAt,
		DiskSize:              *db.DiskSize,
		State:                 string(*db.State),
		TerminationProtection: *db.TerminationProtection,

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
				if db.IpFilter != nil {
					v = *db.IpFilter
				}
				return
			}(),
			URI:        *db.Uri,
			URIParams:  *db.UriParams,
			Version:    defaultString(db.Version, ""),
			Components: components,
			ConnectionInfo: dbServiceOpensearchConnectionInfoShowOutput{
				DashboardURI: *db.ConnectionInfo.DashboardUri,
				Password:     *db.ConnectionInfo.Password,
				URI:          *db.ConnectionInfo.Uri,
				Username:     *db.ConnectionInfo.Username,
			},
			Description:              *db.Description,
			IndexPatterns:            string(indexPatterns),
			IndexTemplate:            indexTemplate,
			KeepIndexRefreshInterval: *db.KeepIndexRefreshInterval,
			MaxIndexCount:            *db.MaxIndexCount,
			Dashboard:                dashboard,
			Users:                    []dbServiceOpensearchUserShowOutput{},
		},
	}, nil
}
