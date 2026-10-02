package acl

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasAclCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_        bool   `cli-cmd:"create"`
	Name     string `cli-arg:"#" cli-usage:"NAME"`
	Username string `cli-arg:"#" cli-usage:"USERNAME"`

	Permission string `cli-usage:"permission to grant (kafka topic: admin|read|readwrite|write, kafka schema registry: schema_registry_read|schema_registry_write, opensearch: admin|read|readwrite|write|deny)"`
	Zone       string `cli-short:"z" cli-usage:"Database Service zone"`

	// "kafka" type specific flags
	KafkaTopic                  string `cli-flag:"kafka-topic" cli-usage:"Kafka topic name or pattern the entry applies to"`
	KafkaSchemaRegistryResource string `cli-flag:"kafka-schema-registry-resource" cli-usage:"Kafka Schema Registry resource name or pattern the entry applies to"`

	// "opensearch" type specific flags
	OpensearchIndex string `cli-flag:"opensearch-index" cli-usage:"OpenSearch index pattern the entry applies to"`
}

func (c *dbaasAclCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *dbaasAclCreateCmd) CmdShort() string     { return "Create an ACL entry for a Database Service" }
func (c *dbaasAclCreateCmd) CmdLong() string {
	return `This command creates an ACL entry for a user of a Kafka or OpenSearch
Database Service.

A Kafka entry applies either to a topic (--kafka-topic) or to a Schema
Registry resource (--kafka-schema-registry-resource). An OpenSearch entry
applies to an index pattern (--opensearch-index).`
}

func (c *dbaasAclCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasAclCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	if c.Permission == "" {
		return errors.New("--permission is required")
	}

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	dbType, err := dbaasServiceType(ctx, client, c.Name, c.Zone)
	if err != nil {
		return err
	}

	switch dbType {
	case "kafka":
		err = c.createKafka(ctx, client)
	case "opensearch":
		err = c.createOpensearch(ctx, client)
	default:
		return fmt.Errorf("creating ACL entries unsupported for service of type %q", dbType)
	}
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc(listAcl(ctx, client, dbType, c.Name))
	}

	return nil
}

func (c *dbaasAclCreateCmd) createKafka(ctx context.Context, client *v3.Client) error {
	var (
		op  *v3.Operation
		err error
	)

	switch {
	case c.KafkaTopic != "" && c.KafkaSchemaRegistryResource != "":
		return errors.New("--kafka-topic and --kafka-schema-registry-resource are mutually exclusive")
	case c.KafkaTopic != "":
		op, err = client.CreateDBAASKafkaTopicAclConfig(ctx, c.Name, v3.DBAASKafkaTopicAclEntry{
			Username:   c.Username,
			Topic:      c.KafkaTopic,
			Permission: v3.DBAASKafkaTopicAclEntryPermission(c.Permission),
		})
	case c.KafkaSchemaRegistryResource != "":
		op, err = client.CreateDBAASKafkaSchemaRegistryAclConfig(ctx, c.Name, v3.DBAASKafkaSchemaRegistryAclEntry{
			Username:   c.Username,
			Resource:   c.KafkaSchemaRegistryResource,
			Permission: v3.DBAASKafkaSchemaRegistryAclEntryPermission(c.Permission),
		})
	default:
		return errors.New("either --kafka-topic or --kafka-schema-registry-resource is required for a kafka service")
	}
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating ACL entry for user %q", c.Username), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func (c *dbaasAclCreateCmd) createOpensearch(ctx context.Context, client *v3.Client) error {
	if c.OpensearchIndex == "" {
		return errors.New("--opensearch-index is required for an opensearch service")
	}

	config, err := client.GetDBAASOpensearchAclConfig(ctx, c.Name)
	if err != nil {
		return err
	}

	rule := v3.DBAASOpensearchAclConfigAclsRules{
		Index:      c.OpensearchIndex,
		Permission: v3.EnumOpensearchRulePermission(c.Permission),
	}

	userFound := false
	for i, acl := range config.Acls {
		if string(acl.Username) != c.Username {
			continue
		}
		userFound = true
		for _, r := range acl.Rules {
			if r.Index == c.OpensearchIndex {
				return fmt.Errorf(
					"an ACL entry already exists for user %q and index %q, use `exo dbaas acl update` to change its permission",
					c.Username,
					c.OpensearchIndex,
				)
			}
		}
		config.Acls[i].Rules = append(config.Acls[i].Rules, rule)
		break
	}
	if !userFound {
		config.Acls = append(config.Acls, v3.DBAASOpensearchAclConfigAcls{
			Username: v3.DBAASUserUsername(c.Username),
			Rules:    []v3.DBAASOpensearchAclConfigAclsRules{rule},
		})
	}

	return updateOpensearchAclConfig(ctx, client, c.Name, *config, fmt.Sprintf("Creating ACL entry for user %q", c.Username))
}

// updateOpensearchAclConfig replaces the ACL configuration of an OpenSearch
// Database Service: the API only exposes it as a whole.
func updateOpensearchAclConfig(ctx context.Context, client *v3.Client, name string, config v3.DBAASOpensearchAclConfig, message string) error {
	op, err := client.UpdateDBAASOpensearchAclConfig(ctx, name, config)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(message, func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasAclCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
