package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-wordwrap"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
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
	res, err := globalstate.EgoscaleClient.GetDbaasServiceOpensearchWithResponse(ctx, oapi.DbaasServiceName(c.Name))
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return nil, fmt.Errorf("resource not found in zone %q", c.Zone)
		}
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
		// Read password from dedicated endpoint
		client, err := switchClientZoneV3(
			ctx,
			globalstate.EgoscaleV3Client,
			v3.ZoneName(c.Zone),
		)
		if err != nil {
			return nil, err
		}

		uriParams := *res.JSON200.UriParams

		creds, err := client.RevealDBAASOpensearchUserPassword(
			ctx,
			string(res.JSON200.Name),
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

func opensearchShowNotifications(db *oapi.DbaasServiceOpensearch) (output.Outputter, error) {
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

func opensearchShowBackups(db *oapi.DbaasServiceOpensearch) (output.Outputter, error) {
	if db.Backups == nil {
		return &dbServiceBackupListOutput{}, nil
	}

	out := make(dbServiceBackupListOutput, 0, len(*db.Backups))
	for _, b := range *db.Backups {
		out = append(out, dbServiceBackupListItemOutput{
			Date: b.BackupTime,
			Name: b.BackupName,
			Size: b.DataSize,
		})
	}

	return &out, nil
}

func opensearchShowDatabase(db *oapi.DbaasServiceOpensearch, zone string) (output.Outputter, error) {
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

	var indexPatterns []dbServiceOpensearchIndexPatternShowOutput
	if db.IndexPatterns != nil {
		for _, i := range *db.IndexPatterns {
			indexPatterns = append(indexPatterns, dbServiceOpensearchIndexPatternShowOutput{
				MaxIndexCount: utils.DefaultInt64(i.MaxIndexCount, 0),
				Pattern:       utils.DefaultString(i.Pattern, ""),
				SortingAlgorithm: func() string {
					if i.SortingAlgorithm != nil {
						return string(*i.SortingAlgorithm)
					}
					return ""
				}(),
			})
		}
	}

	var indexTemplate *dbServiceOpensearchIndexTemplateShowOutput
	if db.IndexTemplate != nil {
		indexTemplate = &dbServiceOpensearchIndexTemplateShowOutput{
			MappingNestedObjectsLimit: utils.DefaultInt64(db.IndexTemplate.MappingNestedObjectsLimit, 0),
			NumberOfReplicas:          utils.DefaultInt64(db.IndexTemplate.NumberOfReplicas, 0),
			NumberOfShards:            utils.DefaultInt64(db.IndexTemplate.NumberOfShards, 0),
		}
	}

	var dashboard *dbServiceOpensearchDashboardShowOutput
	if db.OpensearchDashboards != nil {
		dashboard = &dbServiceOpensearchDashboardShowOutput{
			Enabled:                  utils.DefaultBool(db.OpensearchDashboards.Enabled, false),
			MaxOldSpaceSize:          utils.DefaultInt64(db.OpensearchDashboards.MaxOldSpaceSize, 0),
			OpensearchRequestTimeout: utils.DefaultInt64(db.OpensearchDashboards.OpensearchRequestTimeout, 0),
		}
	}

	var users []dbServiceOpensearchUserShowOutput
	if db.Users != nil {
		for _, u := range *db.Users {
			users = append(users, dbServiceOpensearchUserShowOutput{
				Password: utils.DefaultString(u.Password, ""),
				Type:     utils.DefaultString(u.Type, ""),
				Username: utils.DefaultString(u.Username, ""),
			})
		}
	}

	return &dbServiceShowOutput{
		Zone: zone,
		Name: string(db.Name),
		Type: string(db.Type),
		Plan: db.Plan,
		CreationDate: func() time.Time {
			if db.CreatedAt != nil {
				return *db.CreatedAt
			}
			return time.Time{}
		}(),
		Nodes:      utils.DefaultInt64(db.NodeCount, 0),
		NodeCPUs:   utils.DefaultInt64(db.NodeCpuCount, 0),
		NodeMemory: utils.DefaultInt64(db.NodeMemory, 0),
		UpdateDate: func() time.Time {
			if db.UpdatedAt != nil {
				return *db.UpdatedAt
			}
			return time.Time{}
		}(),
		DiskSize: utils.DefaultInt64(db.DiskSize, 0),
		State: func() string {
			if db.State != nil {
				return string(*db.State)
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
				if db.IpFilter != nil {
					v = *db.IpFilter
				}
				return
			}(),
			URI: utils.DefaultString(db.Uri, ""),
			URIParams: func() map[string]interface{} {
				if db.UriParams != nil {
					return *db.UriParams
				}
				return map[string]interface{}{}
			}(),
			Version:    utils.DefaultString(db.Version, ""),
			Components: components,
			ConnectionInfo: dbServiceOpensearchConnectionInfoShowOutput{
				DashboardURI: utils.DefaultString(db.ConnectionInfo.DashboardUri, ""),
				Password:     utils.DefaultString(db.ConnectionInfo.Password, ""),
				URI: func() []string {
					if db.ConnectionInfo.Uri != nil {
						return *db.ConnectionInfo.Uri
					}
					return []string{}
				}(),
				Username: utils.DefaultString(db.ConnectionInfo.Username, ""),
			},
			Description:              utils.DefaultString(db.Description, ""),
			IndexPatterns:            indexPatterns,
			IndexTemplate:            indexTemplate,
			KeepIndexRefreshInterval: utils.DefaultBool(db.KeepIndexRefreshInterval, false),
			Dashboard:                dashboard,
			Users:                    users,
		},
	}, nil
}
