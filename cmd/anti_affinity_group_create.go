package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type antiAffinityGroupCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Description string `cli-usage:"Anti-Affinity Group description"`
}

func (c *antiAffinityGroupCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *antiAffinityGroupCreateCmd) cmdShort() string {
	return "Create an Anti-Affinity Group"
}

func (c *antiAffinityGroupCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Anti-Affinity Group.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&antiAffinityGroupShowOutput{}), ", "))
}

func (c *antiAffinityGroupCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	antiAffinityGroup := &egoscale.AntiAffinityGroup{
		Description: utils.NonEmptyStringPtr(c.Description),
		Name:        &c.Name,
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Creating Anti-Affinity Group %q...", c.Name), func() {
		antiAffinityGroup, err = globalstate.EgoscaleClient.CreateAntiAffinityGroup(ctx, zone, antiAffinityGroup)
	})
	if err != nil {
		return err
	}

	return (&antiAffinityGroupShowCmd{
		cliCommandSettings: c.cliCommandSettings,
		AntiAffinityGroup:  *antiAffinityGroup.ID,
	}).cmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
