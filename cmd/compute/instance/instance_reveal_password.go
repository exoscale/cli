package instance

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceRevealCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reveal-password"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone     string `cli-short:"z" cli-usage:"instance zone"`
}

type instanceRevealOutput struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

func (o *instanceRevealOutput) Type() string { return "Compute instance" }
func (o *instanceRevealOutput) ToJSON()      { output.JSON(o) }
func (o *instanceRevealOutput) ToText()      { output.Text(o) }
func (o *instanceRevealOutput) ToTable()     { output.Table(o) }

func (c *instanceRevealCmd) CmdAliases() []string { return nil }

func (c *instanceRevealCmd) CmdShort() string { return "Reveal the password of a Compute instance" }

func (c *instanceRevealCmd) CmdLong() string { return "" }

func (c *instanceRevealCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceRevealCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	pwd, err := client.RevealInstancePassword(ctx, instance.ID)
	if err != nil {
		return err
	}

	out := instanceRevealOutput{
		ID:       instance.ID.String(),
		Password: pwd.Password,
	}
	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceRevealCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
