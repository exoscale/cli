package cmd

import (
	"context"
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type dbaasAclDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_           bool   `cli-cmd:"delete"`
	Name        string `cli-flag:"name" cli-usage:"Name of the DBaaS service"`
	ServiceType string `cli-flag:"type" cli-short:"t" cli-usage:"DBaaS service type (e.g., kafka, opensearch)"`
	Username    string `cli-flag:"username" cli-usage:"Username of the ACL entry"`
}

// Command aliases (none in this case)
func (c *dbaasAclDeleteCmd) cmdAliases() []string { return nil }

// Short description for the command
func (c *dbaasAclDeleteCmd) cmdShort() string { return "Delete an ACL entry for a DBaaS service" }

// Long description for the command
func (c *dbaasAclDeleteCmd) cmdLong() string {
	return `This command deletes a specified ACL entry for a DBaaS service, such as Kafka or OpenSearch, across all available zones.`
}

func (c *dbaasAclDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args) // Default validations
}

// Main run logic for showing ACL details
func (c *dbaasAclDeleteCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate required flags
	if c.Name == "" || c.ServiceType == "" || c.Username == "" {
		return fmt.Errorf("all flags --name, --type, and --username must be specified")
	}

	// Search for the service in each zone
	service, zone, err := FindServiceAcrossZones(ctx, globalstate.EgoscaleV3Client, c.Name)
	if err != nil {
		return fmt.Errorf("error finding service: %w", err)
	}

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
	if err != nil {
		return fmt.Errorf("error initializing client for zone %s: %w", zone, err)
	}

	// Validate the service type
	if string(service.Type) != c.ServiceType {
		return fmt.Errorf("mismatched service type: expected %q but got %q for service %q", c.ServiceType, service.Type, c.Name)
	}

	// Call the appropriate delete logic based on the service type
	switch service.Type {
	case "kafka":
		err = c.deleteKafkaACL(ctx, client, c.Name, c.Username)
	default:
		return fmt.Errorf("deleting ACL unsupported for service type %q", service.Type)
	}

	if err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Successfully deleted ACL entry for username %q in service %q.", c.Username, c.Name))
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasAclCmd, &dbaasAclDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
