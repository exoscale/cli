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

	// Search for the service in each zone
	service, zone, err := FindServiceAcrossZones(ctx, globalstate.EgoscaleV3Client, c.Name)
	if err != nil {
		return fmt.Errorf("error finding service: %w", err)
	}

	// Switch client to the appropriate zone
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(zone))
	if err != nil {
		return fmt.Errorf("error initializing client for zone %s: %w", zone, err)
	}

	// Validate the service type
	if string(service.Type) != c.ServiceType {
		return fmt.Errorf("service type mismatch: expected %q but got %q for service %q", c.ServiceType, service.Type, c.Name)
	}

	// Determine the appropriate update logic based on the service type
	switch service.Type {
	case "opensearch":
		return c.updateOpensearch(ctx, client, c.Name)
	default:
		return fmt.Errorf("update ACL unsupported for service type %q", service.Type)
	}
}

func init() {
	cobra.CheckErr(registerCLICommand(dbaasAclCmd, &dbaasAclUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
