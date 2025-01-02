package cmd

import (
	"context"
	"fmt"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type dbaasAclUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_           bool   `cli-cmd:"update"`
	Name        string `cli-flag:"name" cli-usage:"Name of the DBaaS service"`
	Username    string `cli-flag:"username" cli-usage:"Current username of the ACL entry to update"`
	NewUsername string `cli-flag:"new-username" cli-usage:"New username to replace the current one (optional)"`
	ServiceType string `cli-flag:"type" cli-short:"t" cli-usage:"Type of the DBaaS service (e.g., opensearch)"`
	Pattern     string `cli-flag:"pattern" cli-usage:"The pattern for the ACL rule (index* for OpenSearch or topic for Kafka, max 249 characters)"`
	Permission  string `cli-flag:"permission" cli-usage:"Permission to apply (should be one of admin, read, readwrite, write, or deny (only for OpenSearch))"`
}

// Command aliases (none in this case)
func (c *dbaasAclUpdateCmd) cmdAliases() []string { return nil }

// Short description for the command
func (c *dbaasAclUpdateCmd) cmdShort() string {
	return "Update an ACL entry for a DBaaS service"
}

// Long description for the command
func (c *dbaasAclUpdateCmd) cmdLong() string {
	return `This command updates an ACL entry for a specified DBaaS service. You can also update the username with the --new-username flag.`
}

func (c *dbaasAclUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

// Main run logic for showing ACL details
func (c *dbaasAclUpdateCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate required flags
	if c.Name == "" || c.Username == "" || c.ServiceType == "" {
		return fmt.Errorf("both --name, --username, and --type flags must be specified")
	}

	// Fetch all available zones
	zones, err := globalstate.EgoscaleV3Client.ListZones(ctx)
	if err != nil {
		return fmt.Errorf("error fetching zones: %w", err)
	}

	// Iterate through zones to find the service
	var found bool
	var serviceZone string
	var dbType v3.DBAASDatabaseName
	for _, zone := range zones.Zones {
		db, err := dbaasGetV3(ctx, c.Name, string(zone.Name))
		if err == nil {
			dbType = v3.DBAASDatabaseName(db.Type)
			found = true
			serviceZone = string(zone.Name)
			break
		}
	}

	// Handle case where service is not found in any zone
	if !found {
		return fmt.Errorf("service %q not found in any zone", c.Name)
	}

	// Validate the service type
	if string(dbType) != c.ServiceType {
		return fmt.Errorf("service type mismatch: expected %q but got %q for service %q", c.ServiceType, dbType, c.Name)
	}

	// Determine the appropriate update logic based on the service type
	switch dbType {
	case "opensearch":
		return c.updateOpensearch(ctx, serviceZone, c.Name)
	default:
		return fmt.Errorf("update ACL unsupported for service type %q", dbType)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasAclCmd, &dbaasAclUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
