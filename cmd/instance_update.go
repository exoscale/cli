package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	CloudInitFile string            `cli-flag:"cloud-init" cli-short:"c" cli-usage:"instance cloud-init user data configuration file path"`
	Labels        map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	Name          string            `cli-short:"n" cli-usage:"instance name"`
	Zone          string            `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceUpdateCmd) cmdAliases() []string { return nil }

func (c *instanceUpdateCmd) cmdShort() string { return "Update an Instance " }

func (c *instanceUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an Instance .

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instanceUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		instance.Labels = &c.Labels
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		instance.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := getUserDataFromFile(c.CloudInitFile)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instance.UserData = &userData
		updated = true
	}

	if updated {
		decorateAsyncOperation(fmt.Sprintf("Updating instance %q...", c.Instance), func() {
			if err = cs.UpdateInstance(ctx, c.Zone, instance); err != nil {
				return
			}
		})
		if err != nil {
			return err
		}
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
	cobra.CheckErr(registerCLICommand(instanceCmd, &instanceUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
