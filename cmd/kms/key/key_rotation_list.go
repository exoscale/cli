package key

import (
	"os"
	"strconv"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyRotationListOutput struct {
	v3.ListKmsKeyRotationsResponse
}

func (o *keyRotationListOutput) ToJSON() { output.JSON(o) }
func (o *keyRotationListOutput) ToText() { output.Text(o) }
func (o *keyRotationListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"VERSION",
		"ROTATED_AT",
		"AUTOMATIC",
	})

	for _, rotation := range o.Rotations {
		t.Append([]string{
			strconv.Itoa(rotation.Version),
			rotation.RotatedAT.String(),
			strconv.FormatBool(*rotation.Automatic),
		})
	}
}

type keyRotationListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list-rotation"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyRotationListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *keyRotationListCmd) CmdShort() string {
	return "List all the key material versions of a KMS Key."
}

func (c *keyRotationListCmd) CmdLong() string {
	return "List all the key material versions of a KMS Key."
}

func (c *keyRotationListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyRotationListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListKmsKeyRotations(ctx, v3.UUID(c.Key))
	if err != nil {
		return err
	}
	out := keyRotationListOutput{*resp}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyRotationListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
