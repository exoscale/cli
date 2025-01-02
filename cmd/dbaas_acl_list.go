package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasAclListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_           bool   `cli-cmd:"list"`
	Name        string `cli-flag:"name" cli-usage:"Name of the DBaaS service"`
	ServiceType string `cli-flag:"type" cli-short:"t" cli-usage:"Type of the DBaaS service (e.g., kafka, opensearch)"`
}

// Command aliases (none in this case)
func (c *dbaasAclListCmd) cmdAliases() []string { return nil }

// Short description for the command
func (c *dbaasAclListCmd) cmdShort() string { return "List ACL entries for a DBaaS service" }

// Long description for the command
func (c *dbaasAclListCmd) cmdLong() string {
	return `This command lists ACL entries for a specified DBaaS service, including Kafka and OpenSearch, across all available zones.`
}

// Pre-run validation for required flags
func (c *dbaasAclListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args) // Run default validations
}

// Main run logic for listing ACLs
func (c *dbaasAclListCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate required flags
	if c.Name == "" || c.ServiceType == "" {
		return fmt.Errorf("both --name and --type flags must be specified")
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

	// Determine the appropriate listing logic based on the service type
	var output output.Outputter
	switch service.Type {
	case "kafka":
		output, err = c.listKafkaACL(ctx, client, c.Name)
	case "opensearch":
		output, err = c.listOpenSearchACL(ctx, client, c.Name)
	default:
		return fmt.Errorf("listing ACL unsupported for service type %q", service.Type)
	}

	if err != nil {
		return err
	}

	// Output the fetched details
	return c.outputFunc(output, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasAclCmd, &dbaasAclListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
