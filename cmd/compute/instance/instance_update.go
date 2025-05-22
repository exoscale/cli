package instance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	CloudInitFile     string            `cli-flag:"cloud-init" cli-short:"c" cli-usage:"instance cloud-init user data configuration file path"`
	CloudInitCompress bool              `cli-flag:"cloud-init-compress" cli-usage:"compress instance cloud-init user data"`
	Labels            map[string]string `cli-flag:"label" cli-usage:"instance label (format: key=value)"`
	Name              string            `cli-short:"n" cli-usage:"instance name"`
	Protection        bool              `cli-flag:"protection" cli-usage:"delete protection; set --protection=false to disable instance protection"`
	Zone              string            `cli-short:"z" cli-usage:"instance zone"`
	ReverseDNS        string            `cli-usage:"Reverse DNS Domain"`
}

func (c *instanceUpdateCmd) CmdAliases() []string { return nil }

func (c *instanceUpdateCmd) CmdShort() string { return "Update an Instance " }

func (c *instanceUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates an Instance .

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instanceUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updatedInstance, updatedRDNS bool

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		instance.Labels = &c.Labels
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		instance.Name = &c.Name
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		instance.UserData = &userData
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ReverseDNS)) {
		updatedRDNS = true
	}

	if updatedInstance || updatedRDNS || cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Protection)) {
		utils.DecorateAsyncOperation(fmt.Sprintf("Updating instance %q...", c.Instance), func() {
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

			if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Protection)) {
				var client *v3.Client
				client, err = exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
				if err != nil {
					return
				}

				var instanceID v3.UUID
				var op *v3.Operation
				instanceID, err = v3.ParseUUID(*instance.ID)
				if err != nil {
					return
				}
				if c.Protection {
					op, err = client.AddInstanceProtection(ctx, instanceID)
				} else {
					op, err = client.RemoveInstanceProtection(ctx, instanceID)
				}
				if err != nil {
					return
				}
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			}
		})
		if err != nil {
			return err
		}
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
