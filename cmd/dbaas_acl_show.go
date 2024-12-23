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
	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"ACL Entry"})
	defer table.Render()

	table.Append([]string{"Username", o.Username})
	table.Append([]string{"Topic", o.Topic})
	table.Append([]string{"Permission", o.Permission})
}

// Main command for showing ACLs
type dbaasAclShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_           bool   `cli-cmd:"show"`
	Name        string `cli-flag:"name" cli-usage:"Name of the DBaaS service"`
	Username    string `cli-flag:"username" cli-usage:"Username of the ACL entry"`
	ServiceType string `cli-flag:"type" cli-short:"t" cli-usage:"type of the DBaaS service (e.g., kafka, opensearch)"`
	Zone        string `cli-flag:"zone" cli-short:"z" cli-usage:"Database Service zone"`
}

// Command aliases (none in this case)
func (c *dbaasAclShowCmd) cmdAliases() []string { return nil }

// Short description for the command
func (c *dbaasAclShowCmd) cmdShort() string { return "Show the details of an acl" }

// Long description for the command
func (c *dbaasAclShowCmd) cmdLong() string {
	return `This command show an acl entty and its details for a specified DBAAS service.`
}

// Pre-run validation for required flags and default zone setting
func (c *dbaasAclShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)               // Set the default zone if not specified
	return cliCommandDefaultPreRun(c, cmd, args) // Run default validations
}

// Main run logic for showing ACL details
func (c *dbaasAclShowCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := gContext

	// Validate required flags
	if c.Name == "" || c.Username == "" || c.ServiceType == "" {
		return fmt.Errorf("both --name, --username and --type flags must be specified")
	}

	// Fetch DBaaS service details
	db, err := dbaasGetV3(ctx, c.Name, c.Zone)
	if err != nil {
		return fmt.Errorf("error retrieving DBaaS service %q in zone %q: %w", c.Name, c.Zone, err)
	}

	// Validate that the service type matches the expected type
	if string(db.Type) != c.ServiceType {
		return fmt.Errorf("mismatched service type: expected %q but got %q for service %q", c.ServiceType, db.Type, c.Name)
	}

	// Call the appropriate method based on the service type
	var output output.Outputter
	switch db.Type {
	case "kafka":
		output, err = c.showKafka(ctx, c.Name)
	case "opensearch":
		output, err = c.showOpensearch(ctx, c.Name)
	default:
		return fmt.Errorf("listing ACL unsupported for service of type %q", db.Type)
	}

	if err != nil {
		return err
	}

	// Output the fetched details
	return c.outputFunc(output, nil)
}

// Register the command
func init() {
	cobra.CheckErr(registerCLICommand(dbaasAclCmd, &dbaasAclShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
