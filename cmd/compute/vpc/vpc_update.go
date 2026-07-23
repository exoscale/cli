package vpc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	VPC string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`

	Name        string            `cli-usage:"VPC name"`
	Description string            `cli-usage:"VPC description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"VPC label (format: key=value), clearing the labels is possible by passing [=]"`
	Zone        v3.ZoneName       `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcUpdateCmd) CmdAliases() []string { return nil }

func (c *vpcUpdateCmd) CmdShort() string { return "Update a VPC" }

func (c *vpcUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Virtual Private Cloud.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcShowOutput{}), ", "))
}

func (c *vpcUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	entry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	req := v3.UpdateVpcRequest{}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		req.Name = &c.Name
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		req.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		req.Labels = exocmd.ConvertIfSpecialEmptyMap(c.Labels)
		updated = true
	}

	if updated {
		if _, err := client.UpdateVpc(ctx, entry.ID, req); err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&vpcShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			VPC:                entry.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &vpcUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
