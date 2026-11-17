package internal

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
)

type ServiceShow struct {
	Name              string
	Zone              string
	ShowBackups       bool
	ShowNotifications bool
	ShowSettings      string
	ShowURI           bool
}

type DBServiceNotificationListItemOutput struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type dbServiceNotificationListOutput []DBServiceNotificationListItemOutput

func (o *dbServiceNotificationListOutput) ToJSON() { output.JSON(o) }
func (o *dbServiceNotificationListOutput) ToText() { output.Text(o) }
func (o *dbServiceNotificationListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{"Level", "Message"})
	for _, notification := range *o {
		t.Append([]string{
			notification.Level,
			notification.Message,
		})
	}
}

type DBServiceBackupListItemOutput struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
	Size int64     `json:"size"`
}

type dbServiceBackupListOutput []DBServiceBackupListItemOutput

func (o *dbServiceBackupListOutput) ToJSON()  { output.JSON(o) }
func (o *dbServiceBackupListOutput) ToText()  { output.Text(o) }
func (o *dbServiceBackupListOutput) ToTable() { output.Table(o) }

type dbServiceMaintenanceShowOutput struct {
	DOW  string `json:"dow"`
	Time string `json:"time"`
}

type DBServiceShowOutput struct {
	CreationDate          time.Time                       `json:"creation_date"`
	DiskSize              int64                           `json:"disk_size"`
	Maintenance           *dbServiceMaintenanceShowOutput `json:"maintenance"`
	Name                  string                          `json:"name"`
	NodeCPUs              int64                           `json:"node_cpus"`
	NodeMemory            int64                           `json:"node_memory"`
	Nodes                 int64                           `json:"nodes"`
	Plan                  string                          `json:"plan"`
	State                 string                          `json:"state"`
	TerminationProtection bool                            `json:"termination_protection"`
	Type                  string                          `json:"type"`
	UpdateDate            time.Time                       `json:"update_date"`
	Zone                  string                          `json:"zone"`

	Grafana    *dbServiceGrafanaShowOutput    `json:"grafana,omitempty"`
	Kafka      *DBServiceKafkaShowOutput      `json:"kafka,omitempty"`
	Mysql      *DBServiceMysqlShowOutput      `json:"mysql,omitempty"`
	PG         *DBServicePGShowOutput         `json:"pg,omitempty"`
	Valkey     *dbServiceValkeyShowOutput     `json:"valkey,omitempty"`
	Opensearch *DBServiceOpensearchShowOutput `json:"opensearch,omitempty"`
	Thanos     *dbServiceThanosShowOutput     `json:"thanos,omitempty"`
	Clickhouse *dbServiceClickhouseShowOutput `json:"clickhouse,omitempty"`
}

func (o *DBServiceShowOutput) ToJSON() { output.JSON(o) }
func (o *DBServiceShowOutput) ToText() { output.Text(o) }
func (o *DBServiceShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Database Service"})
	defer t.Render()

	t.Append([]string{"Zone", o.Zone})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Type", o.Type})
	t.Append([]string{"Plan", o.Plan})
	t.Append([]string{"Disk Size", humanize.IBytes(uint64(o.DiskSize))})
	t.Append([]string{"State", o.State})
	t.Append([]string{"Creation Date", fmt.Sprint(o.CreationDate)})
	t.Append([]string{"Update Date", fmt.Sprint(o.UpdateDate)})
	t.Append([]string{"Nodes", fmt.Sprint(o.Nodes)})
	t.Append([]string{"Node CPUs", fmt.Sprint(o.NodeCPUs)})
	t.Append([]string{"Node Memory", humanize.IBytes(uint64(o.NodeMemory))})
	t.Append([]string{"Termination Protected", fmt.Sprint(o.TerminationProtection)})

	t.Append([]string{"Maintenance", func() string {
		if o.Maintenance != nil {
			return fmt.Sprintf("%s (%s)", o.Maintenance.DOW, o.Maintenance.Time)
		}
		return "n/a"
	}()})

	switch {
	case o.Grafana != nil:
		formatDatabaseServiceGrafanaTable(t, o.Grafana)
	case o.Kafka != nil:
		formatDatabaseServiceKafkaTable(t, o.Kafka)
	case o.Opensearch != nil:
		formatDatabaseServiceOpensearchTable(t, o.Opensearch)
	case o.Mysql != nil:
		formatDatabaseServiceMysqlTable(t, o.Mysql)
	case o.PG != nil:
		formatDatabaseServicePGTable(t, o.PG)
	case o.Valkey != nil:
		formatDatabaseServiceValkeyTable(t, o.Valkey)
	case o.Thanos != nil:
		formatDatabaseServiceThanosTable(t, o.Thanos)
	case o.Clickhouse != nil:
		formatDatabaseServiceClickhouseTable(t, o.Clickhouse)
	}
}

// redactDatabaseServiceURI returns a redacted version of the URI provided
// (i.e. masks potential password information).
func redactDatabaseServiceURI(u string) string {
	if uri, err := url.Parse(u); err == nil {
		return uri.Redacted()
	}

	return u
}
