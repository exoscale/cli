package key

import (
	"strconv"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyRotationEnableCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"enable-rotation"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	RotationPeriod string      `cli-flag:"rotation-period" cli-short:"r" cli-usage:"number of days for auto rotation period (90 - 2560, default: 365)"`
	Zone           v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyRotationEnableCmd) CmdAliases() []string { return nil }

func (c *keyRotationEnableCmd) CmdShort() string {
	return "Enable the periodic rotation of a KMS Key."
}

func (c *keyRotationEnableCmd) CmdLong() string {
	return "Enable the periodic rotation of a KMS Key."
}

func (c *keyRotationEnableCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyRotationEnableCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	var req v3.EnableKmsKeyRotationRequest
	if cmd.Flags().Changed("rotation-period") {
		n, err := strconv.Atoi(c.RotationPeriod)
		if err != nil {
			return err
		}
		req.RotationPeriod = n
	}

	if _, err := client.EnableKmsKeyRotation(ctx, v3.UUID(c.Key), req); err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&KeyShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Key:                c.Key,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyRotationEnableCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
