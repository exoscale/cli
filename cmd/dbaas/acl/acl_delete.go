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

type dbaasAclDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_        bool   `cli-cmd:"delete"`
	Name     string `cli-arg:"#" cli-usage:"NAME"`
	Username string `cli-arg:"#" cli-usage:"USERNAME"`

	Force      bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Permission string `cli-usage:"only delete the kafka entry granting this permission"`
	Zone       string `cli-short:"z" cli-usage:"Database Service zone"`

	// "kafka" type specific flags
	KafkaTopic                  string `cli-flag:"kafka-topic" cli-usage:"Kafka topic name or pattern of the entry to delete"`
	KafkaSchemaRegistryResource string `cli-flag:"kafka-schema-registry-resource" cli-usage:"Kafka Schema Registry resource name or pattern of the entry to delete"`

	// "opensearch" type specific flags
	OpensearchIndex string `cli-flag:"opensearch-index" cli-usage:"OpenSearch index pattern of the entry to delete (default: all the entries of the user)"`
}

func (c *dbaasAclDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }
func (c *dbaasAclDeleteCmd) CmdShort() string     { return "Delete an ACL entry of a Database Service" }
func (c *dbaasAclDeleteCmd) CmdLong() string {
	return `This command deletes an ACL entry of a user of a Kafka or OpenSearch
Database Service.

A Kafka entry is selected by its topic (--kafka-topic) or its Schema Registry
resource (--kafka-schema-registry-resource), and by --permission if the user
has several entries for it. For an OpenSearch service, all the entries of the
user are deleted unless --opensearch-index is set.`
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

	switch dbType {
	case "kafka":
		return c.deleteKafka(ctx, client)
	case "opensearch":
		return c.deleteOpensearch(ctx, client)
	default:
		return fmt.Errorf("deleting ACL entries unsupported for service of type %q", dbType)
	}
}

func (c *dbaasAclDeleteCmd) confirm(ctx context.Context) bool {
	if c.Force {
		return true
	}

	return utils.AskQuestion(
		ctx,
		fmt.Sprintf(
			"Are you sure you want to delete the ACL of user %q from service %q?",
			c.Username,
			c.Name,
		),
	)
}

func (c *dbaasAclDeleteCmd) deleteKafka(ctx context.Context, client *v3.Client) error {
	if (c.KafkaTopic == "") == (c.KafkaSchemaRegistryResource == "") {
		return errors.New("exactly one of --kafka-topic or --kafka-schema-registry-resource is required for a kafka service")
	}

	acls, err := client.GetDBAASKafkaAclConfig(ctx, c.Name)
	if err != nil {
		return err
	}

	var ids []string
	if c.KafkaTopic != "" {
		for _, acl := range acls.TopicAcl {
			if acl.Username == c.Username && acl.Topic == c.KafkaTopic &&
				(c.Permission == "" || string(acl.Permission) == c.Permission) {
				ids = append(ids, string(acl.ID))
			}
		}
	} else {
		for _, acl := range acls.SchemaRegistryAcl {
			if acl.Username == c.Username && acl.Resource == c.KafkaSchemaRegistryResource &&
				(c.Permission == "" || string(acl.Permission) == c.Permission) {
				ids = append(ids, string(acl.ID))
			}
		}
	}

	switch {
	case len(ids) == 0:
		return fmt.Errorf("no matching ACL entry found for user %q in service %q", c.Username, c.Name)
	case len(ids) > 1:
		return fmt.Errorf("%d ACL entries match for user %q, use --permission to select one", len(ids), c.Username)
	}

	if !c.confirm(ctx) {
		return nil
	}

	var op *v3.Operation
	if c.KafkaTopic != "" {
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

func (c *dbaasAclDeleteCmd) deleteOpensearch(ctx context.Context, client *v3.Client) error {
	config, err := client.GetDBAASOpensearchAclConfig(ctx, c.Name)
	if err != nil {
		return err
	}

	found := false
	acls := make([]v3.DBAASOpensearchAclConfigAcls, 0, len(config.Acls))
	for _, acl := range config.Acls {
		if string(acl.Username) != c.Username {
			acls = append(acls, acl)
			continue
		}

		if c.OpensearchIndex == "" {
			found = true
			continue
		}

		rules := make([]v3.DBAASOpensearchAclConfigAclsRules, 0, len(acl.Rules))
		for _, r := range acl.Rules {
			if r.Index == c.OpensearchIndex {
				found = true
				continue
			}
			rules = append(rules, r)
		}
		if len(rules) > 0 {
			acl.Rules = rules
			acls = append(acls, acl)
		}
	}
	if !found {
		return fmt.Errorf("no matching ACL entry found for user %q in service %q", c.Username, c.Name)
	}

	if !c.confirm(ctx) {
		return nil
	}

	config.Acls = acls
	return updateOpensearchAclConfig(ctx, client, c.Name, *config, fmt.Sprintf("Deleting ACL entry of user %q", c.Username))
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasAclDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
