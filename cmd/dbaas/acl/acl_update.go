package acl

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type dbaasAclUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_        bool   `cli-cmd:"update"`
	Name     string `cli-arg:"#" cli-usage:"NAME"`
	Username string `cli-arg:"?" cli-usage:"USERNAME"`

	Permission string `cli-usage:"new permission of the entry (admin|read|readwrite|write|deny)"`
	Zone       string `cli-short:"z" cli-usage:"Database Service zone"`

	// "opensearch" type specific flags
	OpensearchACLEnabled         bool   `cli-flag:"opensearch-acl-enabled" cli-usage:"enable OpenSearch ACLs (when disabled, authenticated service users have unrestricted access)"`
	OpensearchExtendedACLEnabled bool   `cli-flag:"opensearch-extended-acl-enabled" cli-usage:"enforce index rules in a limited fashion for requests that use the _mget, _msearch and _bulk APIs"`
	OpensearchIndex              string `cli-flag:"opensearch-index" cli-usage:"OpenSearch index pattern of the entry to update"`
}

func (c *dbaasAclUpdateCmd) CmdAliases() []string { return nil }
func (c *dbaasAclUpdateCmd) CmdShort() string {
	return "Update the ACL configuration of a Database Service"
}
func (c *dbaasAclUpdateCmd) CmdLong() string {
	return `This command updates the ACL configuration of an OpenSearch Database
Service: it enables or disables ACLs (--opensearch-acl-enabled,
--opensearch-extended-acl-enabled), and changes the permission of the entry of
a user for an index pattern (USERNAME, --opensearch-index and --permission).

Kafka ACL entries cannot be updated: delete the entry and create a new one.`
}

func (c *dbaasAclUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dbaasAclUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	if (c.Username == "") != (c.OpensearchIndex == "") || (c.Username == "") != (c.Permission == "") {
		return errors.New("USERNAME, --opensearch-index and --permission must be specified together")
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
	case "opensearch":
	case "kafka":
		return errors.New("updating ACL entries unsupported for service of type \"kafka\": delete the entry and create a new one")
	default:
		return fmt.Errorf("updating ACL configuration unsupported for service of type %q", dbType)
	}

	config, err := client.GetDBAASOpensearchAclConfig(ctx, c.Name)
	if err != nil {
		return err
	}

	updated := false

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchACLEnabled)) {
		config.AclEnabled = &c.OpensearchACLEnabled
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.OpensearchExtendedACLEnabled)) {
		config.ExtendedAclEnabled = &c.OpensearchExtendedACLEnabled
		updated = true
	}

	if c.Username != "" {
		found := false
		for i, acl := range config.Acls {
			if string(acl.Username) != c.Username {
				continue
			}
			for j, r := range acl.Rules {
				if r.Index == c.OpensearchIndex {
					config.Acls[i].Rules[j].Permission = v3.EnumOpensearchRulePermission(c.Permission)
					found = true
				}
			}
		}
		if !found {
			return fmt.Errorf(
				"no ACL entry found for user %q and index %q in service %q",
				c.Username,
				c.OpensearchIndex,
				c.Name,
			)
		}
		updated = true
	}

	if !updated {
		return errors.New("nothing to update")
	}

	if err := updateOpensearchAclConfig(ctx, client, c.Name, *config, fmt.Sprintf("Updating ACL configuration of service %q", c.Name)); err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc(showOpensearchAcl(ctx, client, c.Name))
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &dbaasAclUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
