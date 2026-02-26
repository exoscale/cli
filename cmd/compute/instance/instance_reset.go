package instance

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
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
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := findInstance(instances, c.Instance, c.Zone)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to reset instance %q?", c.Instance)) {
			return nil
		}
	}

	request := v3.ResetInstanceRequest{}

	if c.DiskSize > 0 {
		request.DiskSize = c.DiskSize
	}

	if c.Template != "" {

		templates, err := client.ListTemplates(ctx, v3.ListTemplatesWithVisibility(v3.ListTemplatesVisibility(c.TemplateVisibility)))

		if err != nil {
			return err
		}

		template, err := templates.FindTemplate(c.Template)
		if err != nil {
			return fmt.Errorf(
				"no template %q found with visibility %s in zone %s",
				c.Template,
				c.TemplateVisibility,
				c.Zone,
			)
		}

		request.Template = &template
	}

	op, err := client.ResetInstance(ctx, instance.ID, request)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Resetting instance %q...", c.Instance), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           instance.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
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
