package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

type dbaasUserShowOutput struct {
	Username string `json:"username,omitempty"`
	Type     string `json:"type,omitempty"`
	Password string `json:"password,omitempty"`

	// Additional user info for some DBAAS Services
	MySQL *dbaasMysqlUserShowOutput `json:"mysql,omitempty"`
	Kafka *dbaasKafkaUserShowOutput `json:"kafka,omitempty"`
	Redis *dbaasRedisUserShowOutput `json:"redis,omitempty"`
	PG    *dbaasPGUserShowOutput    `json:"pg,omitempty"`
}

func (o *dbaasUserShowOutput) ToJSON() { output.JSON(o) }
func (o *dbaasUserShowOutput) ToText() { output.Text(o) }

func (o *dbaasUserShowOutput) ToTable() {

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Service User"})
	defer t.Render()

	t.Append([]string{"Username", o.Username})
	t.Append([]string{"Type", o.Type})
	t.Append([]string{"Password", o.Password})

	switch {
	case o.MySQL != nil:
		o.MySQL.formatUser(t)
	case o.Kafka != nil:
		o.Kafka.formatUser(t)
	case o.Redis != nil:
		o.Redis.formatUser(t)
	case o.PG != nil:
		o.PG.formatUser(t)
	}

}

type dbaasUserShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasUserShowCmd) cmdAliases() []string { return nil }

func (c *dbaasUserShowCmd) cmdShort() string { return "Show the details of a user" }

func (c *dbaasUserShowCmd) cmdLong() string {
	return `This command show a user and their details for a specified DBAAS service.`
}

func (c *dbaasUserShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)

	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserShowCmd) cmdRun(cmd *cobra.Command, args []string) error {

	ctx := gContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.outputFunc(c.showMysql(ctx))
	case "kafka":
		return c.outputFunc(c.showKafka(ctx))
	case "pg":
		return c.outputFunc(c.showPG(ctx))
	case "opensearch":
		return c.outputFunc(c.showOpensearch(ctx))
	case "redis":
		return c.outputFunc(c.showRedis(ctx))
	case "grafana":
		return c.outputFunc(c.showGrafana(ctx))
	default:
		return fmt.Errorf("Listing users unsupported for service of type %q", db.Type)

	}

}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasUserCmd, &dbaasUserShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
