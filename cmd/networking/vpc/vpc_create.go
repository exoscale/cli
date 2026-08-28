package vpc

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

type vpcCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	Description string            `cli-usage:"VPC description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"VPC label (format: key=value)"`
	Zone        v3.ZoneName       `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *vpcCreateCmd) CmdShort() string { return "Create a VPC" }

func (c *vpcCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Virtual Private Cloud.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcShowOutput{}), ", "))
}

func (c *vpcCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	req := v3.CreateVpcRequest{
		Name:        c.Name,
		Description: c.Description,
	}

	if len(c.Labels) > 0 {
		req.Labels = c.Labels
	}

	op, err := client.CreateVpc(ctx, req)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating VPC %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&vpcShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			VPC:                op.Reference.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &vpcCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
