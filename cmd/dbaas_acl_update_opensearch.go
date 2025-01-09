package cmd

import (
	"context"
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasAclUpdateCmd) updateOpensearch(ctx context.Context, client *v3.Client, serviceName string) error {

	aclsConfig, err := client.GetDBAASOpensearchAclConfig(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("error fetching ACL configuration for service %q: %w", serviceName, err)
	}

	// Ensure ACL entry for the specified username exists
	var updatedAcls []v3.DBAASOpensearchAclConfigAcls
	var updatedEntry *v3.DBAASOpensearchAclConfigAcls
	found := false

	for _, acl := range aclsConfig.Acls {
		if string(acl.Username) == c.Username {
			found = true

			// Update username if --new-username is provided
			newUsername := c.Username
			if c.NewUsername != "" {
				newUsername = c.NewUsername
			}

			updatedEntry = &v3.DBAASOpensearchAclConfigAcls{
				Username: v3.DBAASUserUsername(newUsername),
				Rules: []v3.DBAASOpensearchAclConfigAclsRules{
					{Index: c.Pattern, Permission: v3.EnumOpensearchRulePermission(c.Permission)},
				},
			}
		} else {
			updatedAcls = append(updatedAcls, acl)
		}
	}

	if !found {
		return fmt.Errorf("ACL entry for username %q not found in service %q", c.Username, serviceName)
	}

	if updatedEntry != nil {
		updatedAcls = append(updatedAcls, *updatedEntry)
	}

	// Update the configuration
	aclsConfig.Acls = updatedAcls
	_, err = client.UpdateDBAASOpensearchAclConfig(ctx, serviceName, *aclsConfig)
	if err != nil {
		return fmt.Errorf("error updating ACL configuration for service %q: %w", serviceName, err)
	}

	fmt.Printf("ACL entry for username %q updated successfully in service %q\n", c.Username, serviceName)
	return nil
}
