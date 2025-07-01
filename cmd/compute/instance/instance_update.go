package instance

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/userdata"
	"github.com/exoscale/cli/utils"
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

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := instances.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	updateRequest := v3.UpdateInstanceRequest{}
	updateRDNSRequest := v3.UpdateReverseDNSInstanceRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {

		// To be fixed in the API spec: allow clearing all labels by setting
		// an empty map[string]string (rightnow being omited)
		updateRequest.Labels = c.Labels
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		updateRequest.Name = c.Name
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.CloudInitFile)) {
		userData, err := userdata.GetUserDataFromFile(c.CloudInitFile, c.CloudInitCompress)
		if err != nil {
			return fmt.Errorf("error parsing cloud-init user data: %w", err)
		}
		updateRequest.UserData = userData
		updatedInstance = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.ReverseDNS)) {
		updateRDNSRequest.DomainName = c.ReverseDNS
		updatedRDNS = true
	}

	if updatedInstance || updatedRDNS || cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Protection)) {

		if updatedInstance {
			op, err := client.UpdateInstance(ctx, instance.ID, updateRequest)
			if err != nil {
				return err
			}
			utils.DecorateAsyncOperation(fmt.Sprintf("Updating instance %q...", c.Instance), func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})
			if err != nil {
				return err
			}
		}
		if updatedRDNS {
			op, err := client.UpdateReverseDNSInstance(ctx, instance.ID, updateRDNSRequest)
			if err != nil {
				return err
			}
			utils.DecorateAsyncOperation(fmt.Sprintf("Updating instance reverse DNS %q...", c.Instance), func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})
			if err != nil {
				return err
			}
		}

		if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Protection)) {
			var op *v3.Operation
			var err error
			if c.Protection {
				op, err = client.AddInstanceProtection(ctx, instance.ID)
			} else {
				op, err = client.RemoveInstanceProtection(ctx, instance.ID)
			}
			if err != nil {
				return err
			}

			utils.DecorateAsyncOperation(fmt.Sprintf("Updating instance protection %q to %v...", c.Instance, c.Protection), func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})
			if err != nil {
				return err
			}

		}
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
