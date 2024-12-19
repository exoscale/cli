package cmd

import (
	"context"
	"fmt"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasAclShowCmd) showKafka(ctx context.Context, serviceName string) (output.Outputter, error) {
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return nil, fmt.Errorf("error initializing client for zone %s: %w", c.Zone, err)
	}

	// Fetch Kafka ACLs for the specified service
	acls, err := client.GetDBAASKafkaAclConfig(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("error fetching ACL configuration for service %q: %w", serviceName, err)
	}

	// Search for the specific username in the fetched ACLs
	for _, acl := range acls.TopicAcl {
		if acl.Username == c.Username {
			return &dbaasAclShowOutput{
				Username:   acl.Username,
				Topic:      acl.Topic,
				Permission: string(acl.Permission),
			}, nil
		}
	}

	return nil, fmt.Errorf("ACL entry for username %q not found in service %q", c.Username, serviceName)
}
