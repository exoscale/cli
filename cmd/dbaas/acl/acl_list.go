package acl

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

const (
	aclKindKafkaTopic          = "topic"
	aclKindKafkaSchemaRegistry = "schema-registry"
	aclKindOpensearchIndex     = "index"
)

type dbaasAclListItemOutput struct {
	Username   string `json:"username"`
	Kind       string `json:"kind"`
	Resource   string `json:"resource"`
	Permission string `json:"permission"`
	ID         string `json:"id,omitempty"`
}

type dbaasAclListOutput []dbaasAclListItemOutput

func (o *dbaasAclListOutput) ToJSON()  { output.JSON(o) }
func (o *dbaasAclListOutput) ToText()  { output.Text(o) }
func (o *dbaasAclListOutput) ToTable() { output.Table(o) }

type dbaasAclListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_    bool   `cli-cmd:"list"`
	Name string `cli-arg:"#" cli-usage:"NAME"`
	Zone string `cli-short:"z" cli-usage:"Database Service zone"`
}

func (c *dbaasAclListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *dbaasAclListCmd) CmdShort() string     { return "List ACL entries of a Database Service" }
func (c *dbaasAclListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists the ACL entries of a Kafka or OpenSearch Database Service.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&dbaasAclListItemOutput{}), ", "))
}

func (c *dbaasAclListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasAclListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	dbType, err := dbaasServiceType(ctx, client, c.Name, c.Zone)
	if err != nil {
		return err
	}

	return c.OutputFunc(listAcl(ctx, client, dbType, c.Name))
}

// listAcl returns the ACL entries of a Database Service of type dbType.
func listAcl(ctx context.Context, client *v3.Client, dbType, name string) (output.Outputter, error) {
	switch dbType {
	case "kafka":
		acls, err := client.GetDBAASKafkaAclConfig(ctx, name)
		if err != nil {
			return nil, err
		}
		return kafkaAclListOutput(acls), nil
	case "opensearch":
		config, err := client.GetDBAASOpensearchAclConfig(ctx, name)
		if err != nil {
			return nil, err
		}
		return opensearchAclListOutput(config), nil
	default:
		return nil, fmt.Errorf("listing ACL entries unsupported for service of type %q", dbType)
	}
}

func kafkaAclListOutput(acls *v3.DBAASKafkaAcls) *dbaasAclListOutput {
	out := make(dbaasAclListOutput, 0)
	for _, acl := range acls.TopicAcl {
		out = append(out, dbaasAclListItemOutput{
			Username:   acl.Username,
			Kind:       aclKindKafkaTopic,
			Resource:   acl.Topic,
			Permission: string(acl.Permission),
			ID:         string(acl.ID),
		})
	}
	for _, acl := range acls.SchemaRegistryAcl {
		out = append(out, dbaasAclListItemOutput{
			Username:   acl.Username,
			Kind:       aclKindKafkaSchemaRegistry,
			Resource:   acl.Resource,
			Permission: string(acl.Permission),
			ID:         string(acl.ID),
		})
	}
	return &out
}

func opensearchAclListOutput(config *v3.DBAASOpensearchAclConfig) *dbaasAclListOutput {
	out := make(dbaasAclListOutput, 0)
	for _, acl := range config.Acls {
		for _, rule := range acl.Rules {
			out = append(out, dbaasAclListItemOutput{
				Username:   string(acl.Username),
				Kind:       aclKindOpensearchIndex,
				Resource:   rule.Index,
				Permission: string(rule.Permission),
			})
		}
	}
	return &out
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasAclListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
