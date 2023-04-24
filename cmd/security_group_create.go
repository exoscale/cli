package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type securityGroupCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Description string `cli-usage:"Security Group description"`
}

func (c *securityGroupCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *securityGroupCreateCmd) cmdShort() string {
	return "Create a Security Group"
}

func (c *securityGroupCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Security Group.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	securityGroup := &egoscale.SecurityGroup{
		Description: utils.NonEmptyStringPtr(c.Description),
		Name:        &c.Name,
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Creating Security Group %q...", c.Name), func() {
		securityGroup, err = cs.CreateSecurityGroup(ctx, zone, securityGroup)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return (&securityGroupShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			SecurityGroup:      *securityGroup.ID,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupCmd, &securityGroupCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
