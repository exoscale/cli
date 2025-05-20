package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceTemplateDeleteCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	TemplateID string `cli-arg:"#" cli-usage:"TEMPLATE-ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"template zone"`
}

func (c *instanceTemplateDeleteCmd) CmdAliases() []string { return GRemoveAlias }

func (c *instanceTemplateDeleteCmd) CmdShort() string {
	return "Delete a Compute instance template"
}

func (c *instanceTemplateDeleteCmd) CmdLong() string { return "" }

func (c *instanceTemplateDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	template, err := globalstate.EgoscaleClient.GetTemplate(ctx, c.Zone, c.TemplateID)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to delete template %s (%q)?",
			c.TemplateID,
			*template.Name,
		)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting template %s...", c.TemplateID), func() {
		err = globalstate.EgoscaleClient.DeleteTemplate(ctx, c.Zone, template)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceTemplateCmd, &instanceTemplateDeleteCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
