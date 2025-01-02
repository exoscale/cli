package cmd

import (
	"context"
	"fmt"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type dbaasAclCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_           bool   `cli-cmd:"create"`
	Name        string `cli-flag:"name" cli-usage:"Name of the DBaaS service"`
	Username    string `cli-flag:"username" cli-usage:"Username for the ACL entry"`
	ServiceType string `cli-flag:"type" cli-short:"t" cli-usage:"Type of the DBaaS service (e.g., kafka opensearch)"`
	Pattern     string `cli-flag:"pattern" cli-usage:"The pattern for the ACL rule (index* for OpenSearch or topic for Kafka, max 249 characters)"`
	Permission  string `cli-flag:"permission" cli-usage:"Permission to apply (should be one of admin, read, readwrite, write, or deny (only for OpenSearch))"`
}

// Command aliases (none in this case)
func (c *dbaasAclCreateCmd) cmdAliases() []string { return nil }

// Short description for the command
func (c *dbaasAclCreateCmd) cmdShort() string {
	return "Create an ACL entry for a DBaaS service"
}

// Long description for the command
func (c *dbaasAclCreateCmd) cmdLong() string {
	return `This command creates an ACL entry for a specified DBaaS service, automatically searching for the service across all available zones.`
}

func (c *dbaasAclCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

// Main run logic for showing ACL details
func (c *dbaasAclCreateCmd) cmdRun(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate required inputs
	if c.Name == "" || c.Username == "" || c.ServiceType == "" || c.Permission == "" || c.Pattern == "" {
		return fmt.Errorf("all --name, --username, --type, --permission and --pattern flags must be specified")
	}

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

	switch service.Type {
	case "kafka":
		return c.createKafka(ctx, client, c.Name)
	case "opensearch":
		return c.createOpensearch(ctx, client, c.Name)
	default:
		return fmt.Errorf("create ACL unsupported for service type %q", service.Type)
	}
}
func init() {
	cobra.CheckErr(registerCLICommand(dbaasAclCmd, &dbaasAclCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
