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

	// Fetch all available zones
	zones, err := globalstate.EgoscaleV3Client.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("error fetching zones: %w", err)
	}

	// Iterate through zones to find the service
	var found bool
	var serviceZone string
	var dbType v3.DBAASDatabaseName // Use DBAASDatabaseName for consistency
	for _, zone := range zones.Zones {
		db, err := dbaasGetV3(ctx, c.Name, string(zone.Name))
		if err == nil {
			dbType = v3.DBAASDatabaseName(db.Type) // Save the type for validation
			found = true
			serviceZone = string(zone.Name)
			break
		}
	}

	// Handle case where service is not found in any zone
	if !found {
		return fmt.Errorf("service %q not found in any zone", c.Name)
	}

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(serviceZone))
	if err != nil {
		return fmt.Errorf("error initializing client for zone %s: %w", serviceZone, err)
	}

	// Validate the service type
	if string(dbType) != c.ServiceType {
		return fmt.Errorf("mismatched service type: expected %q but got %q for service %q", c.ServiceType, dbType, c.Name)
	}

	// Call the appropriate delete logic based on the service type
	switch dbType {
	case "kafka":
		err = c.deleteKafkaACL(ctx, client, c.Name, c.Username)
	case "opensearch": //TODO
		//err = c.deleteOpenSearchACL(ctx, client, c.Name, c.Username)
	default:
		return fmt.Errorf("deleting ACL unsupported for service type %q", dbType)
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
