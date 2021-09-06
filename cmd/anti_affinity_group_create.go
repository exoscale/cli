package cmd

import (
	"fmt"
	"strings"

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
		strings.Join(outputterTemplateAnnotations(&antiAffinityGroupShowOutput{}), ", "))
}

func (c *antiAffinityGroupCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	antiAffinityGroup := &egoscale.AntiAffinityGroup{
		Description: func() (v *string) {
			if c.Description != "" {
				v = &c.Description
			}
			return
		}(),
		Name: &c.Name,
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Creating Anti-Affinity Group %q...", c.Name), func() {
		antiAffinityGroup, err = cs.CreateAntiAffinityGroup(ctx, zone, antiAffinityGroup)
	})
	if err != nil {
		return err
	}

	return output(showAntiAffinityGroup(zone, *antiAffinityGroup.ID))
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
