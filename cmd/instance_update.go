package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	CloudInitFile     string            `cli-flag:"cloud-init" cli-short:"c" cli-usage:"instance cloud-init user data configuration file path"`
	CloudInitCompress bool              `cli-flag:"cloud-init-compress" cli-usage:"compress instance cloud-init user data"`
	Labels            map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	Name              string            `cli-short:"n" cli-usage:"instance name"`
	Zone              string            `cli-short:"z" cli-usage:"instance zone"`
	ReverseDNS        string            `cli-usage:"Reverse DNS Domain"`
}

func (c *instanceUpdateCmd) cmdAliases() []string { return nil }

func (c *instanceUpdateCmd) cmdShort() string { return "Update an Instance " }

func (c *instanceUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an Instance .

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instanceUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updatedInstance, updatedRDNS bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		instance.Labels = &c.Labels
		updatedInstance = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		instance.Name = &c.Name
		updatedInstance = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := getUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instance.UserData = &userData
		updatedInstance = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.ReverseDNS)) {
		updatedRDNS = true
	}

	if updatedInstance || updatedRDNS {
		decorateAsyncOperation(fmt.Sprintf("Updating instance %q...", c.Instance), func() {
			if updatedInstance {
				if err = cs.UpdateInstance(ctx, c.Zone, instance); err != nil {
					return
				}
			}

			if updatedRDNS {
				if c.ReverseDNS == "" {
					err = cs.DeleteInstanceReverseDNS(ctx, c.Zone, *instance.ID)
				} else {
					err = cs.UpdateInstanceReverseDNS(ctx, c.Zone, *instance.ID, c.ReverseDNS)
				}
			}
		})

		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
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
