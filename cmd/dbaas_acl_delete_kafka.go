package cmd

import (
	"context"
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
)

// deleteKafkaACL deletes a Kafka ACL entry for the specified username.
func (c *dbaasAclDeleteCmd) deleteKafkaACL(ctx context.Context, client *v3.Client, serviceName, username string) error {
	// Fetch Kafka ACLs for the service
	acls, err := client.GetDBAASKafkaAclConfig(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("error fetching Kafka ACL configuration: %w", err)
	}

	// Find ACL entries for the given username and delete them
	var found bool
	for _, acl := range acls.TopicAcl {
		if acl.Username == username {
			found = true
			// Use the correct delete function to remove the topic ACL
			op, err := client.DeleteDBAASKafkaTopicAclConfig(ctx, serviceName, string(acl.ID))
			if err != nil {
				return fmt.Errorf("error deleting ACL entry %q for topic %q: %w", acl.ID, acl.Topic, err)
			}

			// Wait for the operation to complete (if applicable)
			_, waitErr := client.Wait(ctx, op, v3.OperationStateSuccess)
			if waitErr != nil {
				return fmt.Errorf("error waiting for ACL deletion operation: %w", waitErr)
			}
		}
	}

	if !found {
		return fmt.Errorf("no ACL entry found for username %q in service %q", username, serviceName)
	}

	return nil
}
