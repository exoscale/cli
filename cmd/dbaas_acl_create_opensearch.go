package cmd

import (
	"context"
	"fmt"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasAclCreateCmd) createOpensearch(ctx context.Context, client *v3.Client, serviceName string) error {
	aclsConfig, err := client.GetDBAASOpensearchAclConfig(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("error fetching ACL configuration for service %q: %w", serviceName, err)
	}

	// Check if an entry with the same username already exists
	for _, acl := range aclsConfig.Acls {
		if string(acl.Username) == c.Username {
			return fmt.Errorf("ACL entry for username %q already exists in service %q", c.Username, serviceName)
		}
	}

	// Create a new ACL entry
	newAcl := v3.DBAASOpensearchAclConfigAcls{
		Username: v3.DBAASUserUsername(c.Username),
		Rules: []v3.DBAASOpensearchAclConfigAclsRules{
			{Index: c.Pattern, Permission: v3.EnumOpensearchRulePermission(c.Permission)},
		},
	}

	// Append the new entry to the existing ACLs
	aclsConfig.Acls = append(aclsConfig.Acls, newAcl)

	// Update the configuration with the new entry
	op, err := client.UpdateDBAASOpensearchAclConfig(ctx, serviceName, *aclsConfig)
	if err != nil {
		return fmt.Errorf("error updating ACL configuration for service %q: %w", serviceName, err)
	}

	// Use decorateAsyncOperation to wait for the operation and provide user feedback
	decorateAsyncOperation(fmt.Sprintf("Creating ACL entry for user %q", c.Username), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	fmt.Printf("ACL entry for username %q created successfully in service %q\n", c.Username, serviceName)
	return nil
}
