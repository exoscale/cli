package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceResetCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force              bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	DiskSize           int64  `cli-usage:"disk size to reset the instance to (default: current instance disk size)"`
	Template           string `cli-usage:"template NAME|ID to reset the instance to (default: current instance template)"`
	TemplateVisibility string `cli-usage:"instance template visibility (public|private)"`
	Zone               string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceResetCmd) cmdAliases() []string { return nil }

func (c *instanceResetCmd) cmdShort() string { return "Reset a Compute instance" }

func (c *instanceResetCmd) cmdLong() string {
	return fmt.Sprintf(`This commands resets a Compute instance to a base template state,
and optionally resizes the instance's disk'.

/!\ **************************************************************** /!\
THIS OPERATION EFFECTIVELY WIPES ALL DATA STORED ON THE INSTANCE'S DISK
/!\ **************************************************************** /!\

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceResetCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceResetCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to reset instance %q?", c.Instance)) {
			return nil
		}
	}

	opts := make([]egoscale.ResetInstanceOpt, 0)

	if c.DiskSize > 0 {
		opts = append(opts, egoscale.ResetInstanceWithDiskSize(c.DiskSize))
	}

	var template *egoscale.Template
	if c.Template != "" {
		template, err = cs.FindTemplate(ctx, c.Zone, c.Template, c.TemplateVisibility)
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

	decorateAsyncOperation(fmt.Sprintf("Resetting instance %q...", c.Instance), func() {
		err = cs.ResetInstance(ctx, c.Zone, instance, opts...)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return (&instanceShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceResetCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		TemplateVisibility: defaultTemplateVisibility,
	}))
}
