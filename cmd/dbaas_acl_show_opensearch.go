package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/globalstate"
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

func (o *dbaasAclShowOpensearchOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Field", "Value"})
	defer t.Render()

	t.Append([]string{"ACL Enabled", fmt.Sprintf("%t", o.AclEnabled)})
	t.Append([]string{"Extended ACL Enabled", fmt.Sprintf("%t", o.ExtendedAclEnabled)})

	for _, rule := range o.Rules {
		t.Append([]string{"Rule", fmt.Sprintf("ACL pattern: %s, Permission: %s", rule.Index, rule.Permission)})
	}
}

func (c *dbaasAclShowCmd) showOpensearch(ctx context.Context, serviceName string) (output.Outputter, error) {
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return nil, fmt.Errorf("error initializing client for zone %s: %w", c.Zone, err)
	}

	aclsConfig, err := client.GetDBAASOpensearchAclConfig(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("error fetching ACL configuration for service %q: %w", serviceName, err)
	}

	aclEnabled := false
	if aclsConfig.AclEnabled != nil {
		aclEnabled = *aclsConfig.AclEnabled
	}

	extendedAclEnabled := false
	if aclsConfig.ExtendedAclEnabled != nil {
		extendedAclEnabled = *aclsConfig.ExtendedAclEnabled
	}

	for _, acl := range aclsConfig.Acls {
		if string(acl.Username) == c.Username {
			return &dbaasAclShowOpensearchOutput{
				Username:           string(acl.Username),
				Rules:              acl.Rules,
				AclEnabled:         aclEnabled,
				ExtendedAclEnabled: extendedAclEnabled,
			}, nil
		}
	}

	return nil, fmt.Errorf("ACL entry for username %q not found in service %q", c.Username, serviceName)
}
