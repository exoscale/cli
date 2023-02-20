package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/mitchellh/go-wordwrap"
)

type dbServiceKafkaAuthenticationShowOutput struct {
	Certificate bool `json:"certificate"`
	SASL        bool `json:"sasl"`
}

type dbServiceKafkaACLShowOutput struct {
	ID         string `json:"id"`
	Permission string `json:"permission"`
	Topic      string `json:"topic"`
	Username   string `json:"username"`
}

type dbServiceKafkaComponentShowOutput struct {
	AuthenticationMethod string `json:"authentication_method"`
	Component            string `json:"component"`
	Host                 string `json:"host"`
	Port                 int64  `json:"port"`
	Route                string `json:"route"`
	Usage                string `json:"usage"`
}

type dbServiceKafkaConnectionInfoShowOutput struct {
	AccessCert  *string   `json:"access_cert,omitempty"`
	AccessKey   *string   `json:"access_key,omitempty"`
	Nodes       *[]string `json:"nodes,omitempty"`
	RegistryURI *string   `json:"registry_uri,omitempty"`
	RestURI     *string   `json:"rest_uri,omitempty"`
}

type dbServiceKafkaUserShowOutput struct {
	AccessCert       *string    `json:"access_cert,omitempty"`
	AccessCertExpiry *time.Time `json:"access_cert-expiry,omitempty"`
	AccessKey        *string    `json:"access_key,omitempty"`
	Password         string     `json:"password"`
	Type             string     `json:"type"`
	Username         string     `json:"username"`
}

type dbServiceKafkaShowOutput struct {
	ACL                   []dbServiceKafkaACLShowOutput          `json:"acl"`
	AuthenticationMethods dbServiceKafkaAuthenticationShowOutput `json:"authentication_methods"`
	Components            []dbServiceKafkaComponentShowOutput    `json:"components"`
	ConnectionInfo        dbServiceKafkaConnectionInfoShowOutput `json:"connection_info"`
	IPFilter              []string                               `json:"ip_filter"`
	KafkaConnectEnabled   bool                                   `json:"kafka_connect_enabled"`
	KafkaRESTEnabled      bool                                   `json:"kafka_rest_enabled"`
	SchemaRegistryEnabled bool                                   `json:"schema_registry_enabled"`
	URI                   string                                 `json:"uri"`
	URIParams             map[string]interface{}                 `json:"uri_params"`
	Users                 []dbServiceKafkaUserShowOutput         `json:"users"`
	Version               string                                 `json:"version"`
}

func formatDatabaseServiceKafkaTable(t *table.Table, o *dbServiceKafkaShowOutput) {
	t.Append([]string{"Version", o.Version})
	t.Append([]string{"URI", redactDatabaseServiceURI(o.URI)})
	t.Append([]string{"IP Filter", strings.Join(o.IPFilter, ", ")})

	t.Append([]string{"Authentication Methods", func() string {
		authMethods := make([]string, 0)
		if o.AuthenticationMethods.Certificate {
			authMethods = append(authMethods, "certificate")
		}
		if o.AuthenticationMethods.SASL {
			authMethods = append(authMethods, "SASL")
		}
		return strings.Join(authMethods, ", ")
	}()})

	t.Append([]string{"Kafka Connect Enabled", fmt.Sprint(o.KafkaConnectEnabled)})
	t.Append([]string{"Kafka REST Enabled", fmt.Sprint(o.KafkaRESTEnabled)})
	t.Append([]string{"Schema Registry Enabled", fmt.Sprint(o.SchemaRegistryEnabled)})

	t.Append([]string{"Components", func() string {
		buf := bytes.NewBuffer(nil)
		ct := table.NewEmbeddedTable(buf)
		ct.SetHeader([]string{" "})
		for _, c := range o.Components {
			ct.Append([]string{
				c.Component,
				fmt.Sprintf("%s:%d", c.Host, c.Port),
				"auth:" + c.AuthenticationMethod,
				"route:" + c.Route,
				"usage:" + c.Usage,
			})
		}
		ct.Render()

		return buf.String()
	}()})

	t.Append([]string{"ACL", func() string {
		buf := bytes.NewBuffer(nil)
		at := table.NewEmbeddedTable(buf)
		at.SetHeader([]string{" "})
		for _, acl := range o.ACL {
			at.Append([]string{
				acl.ID,
				"username:" + acl.Username,
				"topic:" + acl.Topic,
				"permission:" + acl.Permission,
			})
		}
		at.Render()

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

func (c *dbaasServiceShowCmd) showDatabaseServiceKafka(ctx context.Context) (outputter, error) {
	serviceRes, err := cs.GetDbaasServiceKafkaWithResponse(ctx, oapi.DbaasServiceName(c.Name))
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return nil, fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return nil, err
	}
	if serviceRes.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API request error: unexpected status %s", serviceRes.Status())
	}
	databaseService := serviceRes.JSON200

	aclRes, err := cs.GetDbaasKafkaAclConfigWithResponse(ctx, oapi.DbaasServiceName(c.Name))
	if err != nil {
		return nil, err
	}
	if aclRes.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("API request error: unexpected status %s", aclRes.Status())
	}
	aclConfig := aclRes.JSON200

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
		case "kafka":
			serviceSettings = databaseService.KafkaSettings
		case "kafka-connect":
			serviceSettings = databaseService.KafkaConnectSettings
		case "kafka-rest":
			serviceSettings = databaseService.KafkaRestSettings
		case "schema-registry":
			serviceSettings = databaseService.SchemaRegistrySettings
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

		Kafka: &dbServiceKafkaShowOutput{
			ACL: func() (v []dbServiceKafkaACLShowOutput) {
				if aclConfig.TopicAcl != nil {
					for _, acl := range *aclConfig.TopicAcl {
						v = append(v, dbServiceKafkaACLShowOutput{
							ID:         string(*acl.Id),
							Permission: string(acl.Permission),
							Topic:      acl.Topic,
							Username:   acl.Username,
						})
					}
				}
				return
			}(),

			AuthenticationMethods: func() (v dbServiceKafkaAuthenticationShowOutput) {
				if databaseService.AuthenticationMethods != nil {
					v.Certificate = utils.DefaultBool(databaseService.AuthenticationMethods.Certificate, false)
					v.SASL = utils.DefaultBool(databaseService.AuthenticationMethods.Sasl, false)
				}
				return
			}(),

			Components: func() (v []dbServiceKafkaComponentShowOutput) {
				if databaseService.Components != nil {
					for _, c := range *databaseService.Components {
						v = append(v, dbServiceKafkaComponentShowOutput{
							Component:            c.Component,
							Host:                 c.Host,
							AuthenticationMethod: utils.DefaultString((*string)(c.KafkaAuthenticationMethod), "-"),
							Port:                 c.Port,
							Route:                string(c.Route),
							Usage:                string(c.Usage),
						})
					}
				}
				return
			}(),

			ConnectionInfo: dbServiceKafkaConnectionInfoShowOutput{
				AccessCert:  databaseService.ConnectionInfo.AccessCert,
				AccessKey:   databaseService.ConnectionInfo.AccessKey,
				Nodes:       databaseService.ConnectionInfo.Nodes,
				RegistryURI: databaseService.ConnectionInfo.RegistryUri,
				RestURI:     databaseService.ConnectionInfo.RestUri,
			},

			IPFilter: func() (v []string) {
				if databaseService.IpFilter != nil {
					v = *databaseService.IpFilter
				}
				return
			}(),

			KafkaConnectEnabled:   utils.DefaultBool(databaseService.KafkaConnectEnabled, false),
			KafkaRESTEnabled:      utils.DefaultBool(databaseService.KafkaRestEnabled, false),
			SchemaRegistryEnabled: utils.DefaultBool(databaseService.SchemaRegistryEnabled, false),

			URI:       *databaseService.Uri,
			URIParams: *databaseService.UriParams,

			Users: func() (v []dbServiceKafkaUserShowOutput) {
				if databaseService.Users != nil {
					for _, u := range *databaseService.Users {
						v = append(v, dbServiceKafkaUserShowOutput{
							AccessCert:       u.AccessCert,
							AccessCertExpiry: u.AccessCertExpiry,
							AccessKey:        u.AccessKey,
							Password:         utils.DefaultString(u.Password, ""),
							Type:             utils.DefaultString(u.Type, ""),
							Username:         utils.DefaultString(u.Username, ""),
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
