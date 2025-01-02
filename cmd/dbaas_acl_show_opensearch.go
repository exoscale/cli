package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasAclShowOpensearchOutput struct {
	Username           string                                 `json:"username,omitempty"`
	Rules              []v3.DBAASOpensearchAclConfigAclsRules `json:"rules,omitempty"`
	AclEnabled         bool                                   `json:"acl_enabled,omitempty"`
	ExtendedAclEnabled bool                                   `json:"extended_acl_enabled,omitempty"`
}

func (o *dbaasAclShowOpensearchOutput) ToJSON() { output.JSON(o) }

func (o *dbaasAclShowOpensearchOutput) ToText() { output.Text(o) }

// ToTable Define table output formatting for OpenSearch
func (o *dbaasAclShowOpensearchOutput) ToTable() {
	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Field", "Value"})
	defer table.Render()

	// Display whether ACL and extended ACL are enabled
	table.Append([]string{"ACL Enabled", fmt.Sprintf("%t", o.AclEnabled)})
	table.Append([]string{"Extended ACL Enabled", fmt.Sprintf("%t", o.ExtendedAclEnabled)})

	// Iterate over rules and display each
	for _, rule := range o.Rules {
		table.Append([]string{"Rule", fmt.Sprintf("ACL pattern: %s, Permission: %s", rule.Index, rule.Permission)})
	}
}

// Fetch OpenSearch ACL configuration and process its details
func (c *dbaasAclShowCmd) showOpensearch(ctx context.Context, client *v3.Client, serviceName string) (output.Outputter, error) {

	// Fetch OpenSearch ACL configuration for the specified service
	aclsConfig, err := client.GetDBAASOpensearchAclConfig(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("error fetching ACL configuration for service %q: %w", serviceName, err)
	}

	// Check if ACLs are enabled
	aclEnabled := false
	if aclsConfig.AclEnabled != nil {
		aclEnabled = *aclsConfig.AclEnabled
	}

	// Check if extended ACLs are enabled
	extendedAclEnabled := false
	if aclsConfig.ExtendedAclEnabled != nil {
		extendedAclEnabled = *aclsConfig.ExtendedAclEnabled
	}

	// Search for the specific username in the fetched ACLs
	for _, acl := range aclsConfig.Acls {
		if string(acl.Username) == c.Username {
			// Return the ACL details for the matched username
			return &dbaasAclShowOpensearchOutput{
				Username:           string(acl.Username),
				Rules:              acl.Rules,
				AclEnabled:         aclEnabled,
				ExtendedAclEnabled: extendedAclEnabled,
			}, nil
		}
	}
	// If no matching username is found, return an error
	return nil, fmt.Errorf("ACL entry for username %q not found in service %q", c.Username, serviceName)
}
