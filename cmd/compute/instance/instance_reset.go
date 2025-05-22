package instance

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceResetCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force              bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	DiskSize           int64  `cli-usage:"disk size to reset the instance to (default: current instance disk size)"`
	Template           string `cli-usage:"template NAME|ID to reset the instance to (default: current instance template)"`
	TemplateVisibility string `cli-usage:"instance template visibility (public|private)"`
	Zone               string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceResetCmd) CmdAliases() []string { return nil }

func (c *instanceResetCmd) CmdShort() string { return "Reset a Compute instance" }

func (c *instanceResetCmd) CmdLong() string {
	return fmt.Sprintf(`This commands resets a Compute instance to a base template state,
and optionally resizes the instance's disk'.

/!\ **************************************************************** /!\
THIS OPERATION EFFECTIVELY WIPES ALL DATA STORED ON THE INSTANCE'S DISK
/!\ **************************************************************** /!\

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "))
}

func (c *instanceResetCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceResetCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to reset instance %q?", c.Instance)) {
			return nil
		}
	}

	opts := make([]egoscale.ResetInstanceOpt, 0)

	if c.DiskSize > 0 {
		opts = append(opts, egoscale.ResetInstanceWithDiskSize(c.DiskSize))
	}

	var template *egoscale.Template
	if c.Template != "" {
		template, err = globalstate.EgoscaleClient.FindTemplate(ctx, c.Zone, c.Template, c.TemplateVisibility)
		if err != nil {
			return fmt.Errorf(
				"no template %q found with visibility %s in zone %s",
				c.Template,
				c.TemplateVisibility,
				c.Zone,
			)
		}
		opts = append(opts, egoscale.ResetInstanceWithTemplate(template))
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Resetting instance %q...", c.Instance), func() {
		err = globalstate.EgoscaleClient.ResetInstance(ctx, c.Zone, instance, opts...)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceResetCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),

		TemplateVisibility: exocmd.DefaultTemplateVisibility,
	}))
}
