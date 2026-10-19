package acl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasAclDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_        bool   `cli-cmd:"delete"`
	Name     string `cli-arg:"#" cli-usage:"NAME"`
	Username string `cli-arg:"#" cli-usage:"USERNAME"`

	Force      bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	ID         string `cli-usage:"ID of the entry to delete, as shown by \"exo dbaas acl list\""`
	Permission string `cli-usage:"only delete the entry granting this permission"`
	Zone       string `cli-short:"z" cli-usage:"Database Service zone"`

	// "kafka" type specific flags
	KafkaTopic                  string `cli-flag:"kafka-topic" cli-usage:"Kafka topic name or pattern of the entry to delete"`
	KafkaSchemaRegistryResource string `cli-flag:"kafka-schema-registry-resource" cli-usage:"Kafka Schema Registry resource name or pattern of the entry to delete"`
}

func (c *dbaasAclDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }
func (c *dbaasAclDeleteCmd) CmdShort() string     { return "Delete an ACL entry of a Database Service" }
func (c *dbaasAclDeleteCmd) CmdLong() string {
	return `This command deletes an ACL entry of a user of a Kafka Database Service.

The entry is selected by its ID (--id, as shown by "exo dbaas acl list"), or
by its topic (--kafka-topic) or its Schema Registry resource
(--kafka-schema-registry-resource), and by --permission if the user has
several entries for it.`
}

func (c *dbaasAclDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasAclDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	dbType, err := dbaasServiceType(ctx, client, c.Name, c.Zone)
	if err != nil {
		return err
	}

	if dbType != "kafka" {
		return fmt.Errorf("deleting ACL entries unsupported for service of type %q", dbType)
	}

	return c.deleteKafka(ctx, client)
}

func (c *dbaasAclDeleteCmd) confirm(ctx context.Context) bool {
	if c.Force {
		return true
	}

	return utils.AskQuestion(
		ctx,
		fmt.Sprintf(
			"Are you sure you want to delete the ACL entry of user %q from service %q?",
			c.Username,
			c.Name,
		),
	)
}

func (c *dbaasAclDeleteCmd) deleteKafka(ctx context.Context, client *v3.Client) error {
	if c.ID != "" {
		if c.KafkaTopic != "" || c.KafkaSchemaRegistryResource != "" || c.Permission != "" {
			return errors.New("--id can't be used with --kafka-topic, --kafka-schema-registry-resource or --permission")
		}
	} else if (c.KafkaTopic == "") == (c.KafkaSchemaRegistryResource == "") {
		return errors.New("exactly one of --id, --kafka-topic or --kafka-schema-registry-resource is required for a kafka service")
	}

	acls, err := client.GetDBAASKafkaAclConfig(ctx, c.Name)
	if err != nil {
		return err
	}

	var topicIDs, schemaRegistryIDs []string
	for _, acl := range acls.TopicAcl {
		if acl.Username == c.Username && (c.ID != "" && string(acl.ID) == c.ID ||
			c.KafkaTopic != "" && acl.Topic == c.KafkaTopic &&
				(c.Permission == "" || string(acl.Permission) == c.Permission)) {
			topicIDs = append(topicIDs, string(acl.ID))
		}
	}
	for _, acl := range acls.SchemaRegistryAcl {
		if acl.Username == c.Username && (c.ID != "" && string(acl.ID) == c.ID ||
			c.KafkaSchemaRegistryResource != "" && acl.Resource == c.KafkaSchemaRegistryResource &&
				(c.Permission == "" || string(acl.Permission) == c.Permission)) {
			schemaRegistryIDs = append(schemaRegistryIDs, string(acl.ID))
		}
	}

	ids := append(append([]string{}, topicIDs...), schemaRegistryIDs...)
	switch {
	case len(ids) == 0:
		return fmt.Errorf("no matching ACL entry found for user %q in service %q", c.Username, c.Name)
	case len(ids) > 1:
		return fmt.Errorf("%d ACL entries match for user %q (%s), use --permission or --id to select one",
			len(ids), c.Username, strings.Join(ids, ", "))
	}

	if !c.confirm(ctx) {
		return nil
	}

	var op *v3.Operation
	if len(topicIDs) == 1 {
		op, err = client.DeleteDBAASKafkaTopicAclConfig(ctx, c.Name, ids[0])
	} else {
		op, err = client.DeleteDBAASKafkaSchemaRegistryAclConfig(ctx, c.Name, ids[0])
	}
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting ACL entry of user %q", c.Username), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasAclDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
