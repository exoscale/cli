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

	Permission string `cli-usage:"permission to grant (kafka topic: admin|read|readwrite|write, kafka schema registry: schema_registry_read|schema_registry_write)"`
	Zone       string `cli-short:"z" cli-usage:"Database Service zone"`

	// "kafka" type specific flags
	KafkaTopic                  string `cli-flag:"kafka-topic" cli-usage:"Kafka topic name or pattern the entry applies to"`
	KafkaSchemaRegistryResource string `cli-flag:"kafka-schema-registry-resource" cli-usage:"Kafka Schema Registry resource name or pattern the entry applies to"`
}

func (c *dbaasAclCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *dbaasAclCreateCmd) CmdShort() string     { return "Create an ACL entry for a Database Service" }
func (c *dbaasAclCreateCmd) CmdLong() string {
	return `This command creates an ACL entry for a user of a Kafka Database Service.

An entry applies either to a topic (--kafka-topic) or to a Schema Registry
resource (--kafka-schema-registry-resource).`
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

	if dbType != "kafka" {
		return fmt.Errorf("creating ACL entries unsupported for service of type %q", dbType)
	}

	if err := c.createKafka(ctx, client); err != nil {
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

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasAclCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
