package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

type dbaasAclShowOutput struct {
	Username   string `json:"username,omitempty"`
	Permission string `json:"permission,omitempty"`
	Topic      string `json:"topic,omitempty"`
}

func (o *dbaasAclShowOutput) ToJSON() { output.JSON(o) }
func (o *dbaasAclShowOutput) ToText() { output.Text(o) }

func (o *dbaasAclShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"ACL Entry"})
	defer t.Render()

	t.Append([]string{"Username", o.Username})
	t.Append([]string{"Topic", o.Topic})
	t.Append([]string{"Permission", o.Permission})
}

type dbaasAclShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_           bool   `cli-cmd:"show"`
	Name        string `cli-flag:"name" cli-usage:"Name of the DBaaS service"`
	Username    string `cli-flag:"username" cli-usage:"Username of the ACL entry"`
	ServiceType string `cli-short:"t" cli-usage:"type of the DBaaS service (e.g., kafka, opensearch)"`
	Zone        string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasAclShowCmd) cmdAliases() []string { return nil }

func (c *dbaasAclShowCmd) cmdShort() string { return "Show the details of an acl" }

func (c *dbaasAclShowCmd) cmdLong() string {
	return `This command show an acl entty and its details for a specified DBAAS service.`
}

func (c *dbaasAclShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)

	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasAclShowCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := gContext

	if c.Name == "" || c.Username == "" {
		return fmt.Errorf("both --name and --username flags must be specified")
	}

	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return fmt.Errorf("error retrieving DBaaS service %q in zone %q: %w", c.Name, c.Zone, err)
	}

	var output output.Outputter
	switch db.Type {
	case "kafka":
		output, err = c.showKafka(ctx, c.Name)
		//case "opensearch":
		//output, err = c.showOpensearch(ctx)
	default:
		return fmt.Errorf("listing ACL unsupported for service of type %q", db.Type)
	}

	if err != nil {
		return err
	}

	return c.outputFunc(output, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasAclCmd, &dbaasAclShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
