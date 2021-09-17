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
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *instanceResetCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
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

	var template *egoscale.Template
	if c.Template != "" {
		templates, err := cs.ListTemplates(ctx, c.Zone, c.TemplateVisibility, "")
		if err != nil {
			return fmt.Errorf("error retrieving templates: %s", err)
		}
		for _, t := range templates {
			if *t.ID == c.Template || *t.Name == c.Template {
				template = t
				break
			}
		}
		if template == nil {
			return fmt.Errorf("no template %q found with visibility %s", c.Template, c.TemplateVisibility)
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Reseting instance %q...", c.Instance), func() {
		err = cs.ResetInstance(ctx, c.Zone, instance, template, c.DiskSize)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showInstance(c.Zone, *instance.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceResetCmd{
		cliCommandSettings: defaultCLICmdSettings(),

		TemplateVisibility: defaultTemplateVisibility,
	}))
}
