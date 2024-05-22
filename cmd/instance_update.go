package cmd

import (
	"errors"
	"fmt"
	egoscale3 "github.com/exoscale/egoscale/v3"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	CloudInitFile     string            `cli-flag:"cloud-init" cli-short:"c" cli-usage:"instance cloud-init user data configuration file path"`
	CloudInitCompress bool              `cli-flag:"cloud-init-compress" cli-usage:"compress instance cloud-init user data"`
	Labels            map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	Name              string            `cli-short:"n" cli-usage:"instance name"`
	Protection        bool              `cli-flag:"protection" cli-usage:"enable delete protection"`
	Zone              string            `cli-short:"z" cli-usage:"instance zone"`
	ReverseDNS        string            `cli-usage:"Reverse DNS Domain"`
}

func (c *instanceUpdateCmd) cmdAliases() []string { return nil }

func (c *instanceUpdateCmd) cmdShort() string { return "Update an Instance " }

func (c *instanceUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates an Instance .

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instanceUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updatedInstance, updatedRDNS bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
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
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
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
				if err = globalstate.EgoscaleClient.UpdateInstance(ctx, c.Zone, instance); err != nil {
					return
				}
			}

			if updatedRDNS {
				if c.ReverseDNS == "" {
					err = globalstate.EgoscaleClient.DeleteInstanceReverseDNS(ctx, c.Zone, *instance.ID)
				} else {
					err = globalstate.EgoscaleClient.UpdateInstanceReverseDNS(ctx, c.Zone, *instance.ID, c.ReverseDNS)
				}
			}
			var value egoscale3.UUID
			var op *egoscale3.Operation
			value, err = egoscale3.ParseUUID(*instance.ID)
			if err != nil {
				return
			}
			if c.Protection {
				op, err = globalstate.EgoscaleV3Client.AddInstanceProtection(ctx, value)
			} else {
				op, err = globalstate.EgoscaleV3Client.RemoveInstanceProtection(ctx, value)
			}
			if err != nil {
				return
			}
			op, err = globalstate.EgoscaleV3Client.Wait(ctx, op, egoscale3.OperationStateSuccess)

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
