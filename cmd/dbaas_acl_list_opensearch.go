package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

// dbaasAclListOpenSearchOutput defines the OpenSearch ACL output structure.
type dbaasAclListOpenSearchOutput struct {
	Acls               []v3.DBAASOpensearchAclConfigAcls `json:"acls,omitempty"`                 // ACLs grouped by username
	AclEnabled         bool                              `json:"acl_enabled,omitempty"`          // ACL Enabled status
	ExtendedAclEnabled bool                              `json:"extended_acl_enabled,omitempty"` // Extended ACL Enabled status
}

// ToJSON outputs the result in JSON format.
func (o *dbaasAclListOpenSearchOutput) ToJSON() { output.JSON(o) }

// ToText outputs the result in plain text format.
func (o *dbaasAclListOpenSearchOutput) ToText() { output.Text(o) }

// ToTable outputs the result in a tabular format.
func (o *dbaasAclListOpenSearchOutput) ToTable() {
	tabular := table.NewTable(os.Stdout)
	tabular.SetHeader([]string{"Field", "Value"})

	// Display general ACL configurations.
	tabular.Append([]string{"ACL Enabled", fmt.Sprintf("%t", o.AclEnabled)})
	tabular.Append([]string{"Extended ACL Enabled", fmt.Sprintf("%t", o.ExtendedAclEnabled)})

	// Display rules grouped under usernames.
	for _, acl := range o.Acls {
		tabular.Append([]string{"Username", string(acl.Username)})
		for _, rule := range acl.Rules {
			tabular.Append([]string{"  Rule", fmt.Sprintf("Pattern: %s, Permission: %s", rule.Index, string(rule.Permission))})
		}
	}

	tabular.Render()
}

// listOpenSearchACL fetches OpenSearch ACLs and prepares the output.
func (c *dbaasAclListCmd) listOpenSearchACL(ctx context.Context, client *v3.Client, serviceName string) (output.Outputter, error) {
	aclsConfig, err := client.GetDBAASOpensearchAclConfig(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("error fetching OpenSearch ACL configuration: %w", err)
	}

	return &dbaasAclListOpenSearchOutput{
		Acls:               aclsConfig.Acls,
		AclEnabled:         *aclsConfig.AclEnabled,
		ExtendedAclEnabled: *aclsConfig.ExtendedAclEnabled,
	}, nil
}
