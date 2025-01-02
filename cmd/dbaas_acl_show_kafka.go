package cmd

import (
	"context"
	"fmt"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

// Fetch OpenSearch ACL configuration and process its details
func (c *dbaasAclShowCmd) showKafka(ctx context.Context, client *v3.Client, serviceName string) (output.Outputter, error) {

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
