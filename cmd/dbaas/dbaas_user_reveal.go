package dbaas

import (
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

type dbaasUserRevealOutput struct {
	Username string `json:"username,omitempty"`
	Type     string `json:"type,omitempty"`
	Password string `json:"password,omitempty"`

	// Additional user info for some DBAAS Services
	Kafka *dbaasKafkaUserRevealOutput `json:"kafka,omitempty"`
}

func (o *dbaasUserRevealOutput) ToJSON() { output.JSON(o) }
func (o *dbaasUserRevealOutput) ToText() { output.Text(o) }

func (o *dbaasUserRevealOutput) ToTable() {

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Secrets"})
	defer t.Render()

	t.Append([]string{"Password", o.Password})

	switch {
	case o.Kafka != nil:
		o.Kafka.formatUser(t)
	}

}

type dbaasUserRevealCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reveal-secrets"`

	Name     string `cli-arg:"#"`
	Username string `cli-arg:"#"`
	Zone     string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasUserRevealCmd) CmdAliases() []string { return nil }

func (c *dbaasUserRevealCmd) CmdShort() string { return "Show the secrets of a user" }

func (c *dbaasUserRevealCmd) CmdLong() string {
	return `This command reveals a user's password and other possible secrets, depending on the service type.`
}

func (c *dbaasUserRevealCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)

	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasUserRevealCmd) CmdRun(cmd *cobra.Command, args []string) error {

	ctx := exocmd.GContext
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch db.Type {
	case "mysql":
		return c.OutputFunc(c.revealMysql(ctx))
	case "kafka":
		return c.OutputFunc(c.revealKafka(ctx))
	case "pg":
		return c.OutputFunc(c.revealPG(ctx))
	case "opensearch":
		return c.OutputFunc(c.revealOpensearch(ctx))
	case "grafana":
		return c.OutputFunc(c.revealGrafana(ctx))
	case "valkey":
		return c.OutputFunc(c.revealValkey(ctx))
	case "thanos":
		return c.OutputFunc(c.revealThanos(ctx))
	default:
		return fmt.Errorf("listing users unsupported for service of type %q", db.Type)

	}

}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(dbaasUserCmd, &dbaasUserRevealCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
