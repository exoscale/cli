package dbaas

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	dbaasutils "github.com/exoscale/cli/cmd/dbaas/internal"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/output"
)

type dbaasServiceShowCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#"`

	ShowBackups       bool   `cli-flag:"backups" cli-usage:"show Database Service backups"`
	ShowNotifications bool   `cli-flag:"notifications" cli-usage:"show Database Service notifications"`
	ShowSettings      string `cli-flag:"settings" cli-usage:"show Database Service settings (see \"exo dbaas type show --help\" for supported settings)"`
	ShowURI           bool   `cli-flag:"uri" cli-usage:"show Database Service connection URI"`
	Zone              string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasServiceShowCmd) CmdAliases() []string { return clicmd.GShowAlias }

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
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceKafkaShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceKafkaACLShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceKafkaAuthenticationShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceKafkaComponentShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceKafkaConnectionInfoShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceKafkaUserShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceOpensearchShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceMysqlShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceMysqlComponentShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceMysqlUserShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServicePGShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServicePGComponentShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServicePGConnectionPool{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServicePGUserShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceBackupListItemOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&dbaasutils.DBServiceNotificationListItemOutput{}), ", "))
}

func (c *dbaasServiceShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	clicmd.CmdSetZoneFlagFromDefault(cmd)
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasServiceShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	var err error

	svc, err := dbaasutils.GetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	s := &dbaasutils.ServiceShow{
		Name:              c.Name,
		Zone:              c.Zone,
		ShowBackups:       c.ShowBackups,
		ShowNotifications: c.ShowNotifications,
		ShowSettings:      c.ShowSettings,
		ShowURI:           c.ShowURI,
	}

	switch svc.Type {
	case "grafana":
		return c.OutputFunc(s.ShowDatabaseServiceGrafana(ctx))
	case "kafka":
		return c.OutputFunc(s.ShowDatabaseServiceKafka(ctx))
	case "opensearch":
		return c.OutputFunc(s.ShowDatabaseServiceOpensearch(ctx))
	case "mysql":
		return c.OutputFunc(s.ShowDatabaseServiceMysql(ctx))
	case "pg":
		return c.OutputFunc(s.ShowDatabaseServicePG(ctx))
	case "valkey":
		return c.OutputFunc(s.ShowDatabaseServiceValkey(ctx))
	case "thanos":
		return c.OutputFunc(s.ShowDatabaseServiceThanos(ctx))
	case "clickhouse":
		return c.OutputFunc(s.ShowDatabaseServiceClickhouse(ctx))
	default:
		return fmt.Errorf("unsupported service type %q", svc.Type)
	}
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dbaasServiceShowCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
