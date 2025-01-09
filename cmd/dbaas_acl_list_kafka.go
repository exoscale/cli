package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

// dbaasAclListKafkaOutput defines the Kafka ACL output structure.
type dbaasAclListKafkaOutput struct {
	TopicAcls []v3.DBAASKafkaTopicAclEntry `json:"topic_acls,omitempty"`
}

// ToJSON outputs the result in JSON format.
func (o *dbaasAclListKafkaOutput) ToJSON() { output.JSON(o) }

// ToText outputs the result in plain text format.
func (o *dbaasAclListKafkaOutput) ToText() { output.Text(o) }

// ToTable outputs the result in a tabular format.
func (o *dbaasAclListKafkaOutput) ToTable() {
	tabular := table.NewTable(os.Stdout)
	tabular.SetHeader([]string{"Username", "Topic", "Permission"})

	// Display Topic ACL entries.
	for _, acl := range o.TopicAcls {
		tabular.Append([]string{acl.Username, acl.Topic, string(acl.Permission)})
	}

	tabular.Render()
}

// listKafkaACL fetches Kafka ACLs and prepares the output.
func (c *dbaasAclListCmd) listKafkaACL(ctx context.Context, client *v3.Client, serviceName string) (output.Outputter, error) {
	acls, err := client.GetDBAASKafkaAclConfig(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("error fetching Kafka ACL configuration: %w", err)
	}

	return &dbaasAclListKafkaOutput{
		TopicAcls: acls.TopicAcl,
	}, nil
}
