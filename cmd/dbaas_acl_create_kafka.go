package cmd

import (
	"context"
	"fmt"

	v3 "github.com/exoscale/egoscale/v3"
)

func (c *dbaasAclCreateCmd) createKafka(ctx context.Context, client *v3.Client, serviceName string) error {
	// Define the new Kafka ACL entry
	newAcl := v3.DBAASKafkaTopicAclEntry{
		Username:   c.Username,
		Topic:      c.Pattern,
		Permission: v3.DBAASKafkaTopicAclEntryPermission(c.Permission),
	}

	// Trigger the creation of the ACL entry
	op, err := client.CreateDBAASKafkaTopicAclConfig(ctx, serviceName, newAcl)
	if err != nil {
		return fmt.Errorf("error creating ACL entry for service %q: %w", serviceName, err)
	}

	// Use decorateAsyncOperation to handle the operation and provide user feedback
	decorateAsyncOperation(fmt.Sprintf("Creating Kafka ACL entry for user %q", c.Username), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	if err != nil {
		return fmt.Errorf("error completing ACL creation: %w", err)
	}

	fmt.Printf("Kafka ACL entry for user %q successfully created in service %q\n", c.Username, serviceName)
	return nil
}
