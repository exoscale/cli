package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceTemplateDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	TemplateID string `cli-arg:"#" cli-usage:"TEMPLATE-ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"template zone"`
}

func (c *instanceTemplateDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *instanceTemplateDeleteCmd) CmdShort() string {
	return "Delete a Compute instance template"
}

func (c *instanceTemplateDeleteCmd) CmdLong() string { return "" }

func (c *instanceTemplateDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTemplateDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	template, err := globalstate.EgoscaleClient.GetTemplate(ctx, c.Zone, c.TemplateID)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to delete template %s (%q)?",
				c.TemplateID,
				*template.Name,
			)) {
			return nil
		}
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting template %s...", c.TemplateID), func() {
		err = globalstate.EgoscaleClient.DeleteTemplate(ctx, c.Zone, template)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceTemplateCmd, &instanceTemplateDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
