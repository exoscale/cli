package dbaas

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
)

type dbServiceNotificationListItemOutput struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type dbServiceNotificationListOutput []dbServiceNotificationListItemOutput

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

type dbServiceBackupListItemOutput struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
	Size int64     `json:"size"`
}

type dbServiceBackupListOutput []dbServiceBackupListItemOutput

func (o *dbServiceBackupListOutput) ToJSON()  { output.JSON(o) }
func (o *dbServiceBackupListOutput) ToText()  { output.Text(o) }
func (o *dbServiceBackupListOutput) ToTable() { output.Table(o) }

type dbServiceMaintenanceShowOutput struct {
	DOW  string `json:"dow"`
	Time string `json:"time"`
}

type dbServiceShowOutput struct {
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
	Kafka      *dbServiceKafkaShowOutput      `json:"kafka,omitempty"`
	Mysql      *dbServiceMysqlShowOutput      `json:"mysql,omitempty"`
	PG         *dbServicePGShowOutput         `json:"pg,omitempty"`
	Valkey     *dbServiceValkeyShowOutput     `json:"valkey,omitempty"`
	Opensearch *dbServiceOpensearchShowOutput `json:"opensearch,omitempty"`
}

func (o *dbServiceShowOutput) ToJSON() { output.JSON(o) }
func (o *dbServiceShowOutput) ToText() { output.Text(o) }
func (o *dbServiceShowOutput) ToTable() {
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
	}
}

type dbaasServiceShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#"`

	ShowBackups       bool   `cli-flag:"backups" cli-usage:"show Database Service backups"`
	ShowNotifications bool   `cli-flag:"notifications" cli-usage:"show Database Service notifications"`
	ShowSettings      string `cli-flag:"settings" cli-usage:"show Database Service settings (see \"exo dbaas type show --help\" for supported settings)"`
	ShowURI           bool   `cli-flag:"uri" cli-usage:"show Database Service connection URI"`
	Zone              string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *dbaasServiceShowCmd) CmdShort() string { return "Show a Database Service details" }

func (c *dbaasServiceShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Database Service details.

Supported output template annotations:

* When showing a Database Service: %s
  - .Kafka: %s
    - .Kafka.ACL[]: %s
    - .Kafka.AuthenticationMethods: %s
    - .Kafka.Components[]: %s
    - .Kafka.ConnectionInfo: %s
    - .Kafka.Users[]: %s
  - .Opensearch: %s
  - .Mysql: %s
    - .Mysql.Components[]: %s
    - .Mysql.Users[]: %s
  - .PG: %s
    - .PG.Components[]: %s
    - .PG.ConnectionPools: %s
    - .PG.Users[]: %s

* When showing a Database Service backups: %s

* When showing a Database Service notifications: %s`,
		strings.Join(output.TemplateAnnotations(&dbServiceShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceKafkaShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceKafkaACLShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceKafkaAuthenticationShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceKafkaComponentShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceKafkaConnectionInfoShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceKafkaUserShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceOpensearchShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceMysqlShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceMysqlComponentShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceMysqlUserShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServicePGShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServicePGComponentShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServicePGConnectionPool{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServicePGUserShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceBackupListItemOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbServiceNotificationListItemOutput{}), ", "))
}

func (c *dbaasServiceShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	var err error

	svc, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch svc.Type {
	case "grafana":
		return c.OutputFunc(c.showDatabaseServiceGrafana(ctx))
	case "kafka":
		return c.OutputFunc(c.showDatabaseServiceKafka(ctx))
	case "opensearch":
		return c.OutputFunc(c.showDatabaseServiceOpensearch(ctx))
	case "mysql":
		return c.OutputFunc(c.showDatabaseServiceMysql(ctx))
	case "pg":
		return c.OutputFunc(c.showDatabaseServicePG(ctx))
	case "valkey":
		return c.OutputFunc(c.showDatabaseServiceValkey(ctx))
	default:
		return fmt.Errorf("unsupported service type %q", svc.Type)
	}
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasCmd, &dbaasServiceShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
