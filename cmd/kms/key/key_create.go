package key

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	Description string      `cli-short:"d" cli-flag:"description" cli-usage:"key description"`
	Usage       string      `cli-short:"u" cli-flag:"usage" cli-usage:"key usage [encrypt-decrypt]"`
	Multizone   bool        `cli-short:"m" cli-flag:"multizone" cli-usage:"allow replication accross zones (default: false)"`
	Zone        v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyCreateCmd) CmdAliases() []string { return nil }

func (c *keyCreateCmd) CmdShort() string {
	return "Create a KMS Key in a given zone with a given name."
}

func (c *keyCreateCmd) CmdLong() string {
	return "Create a KMS Key in a given zone with a given name."
}

func (c *keyCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyCreateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	req := v3.CreateKmsKeyRequest{
		Name: c.Name,
	}

	if cmd.Flags().Changed("usage") {
		req.Usage = v3.CreateKmsKeyRequestUsage(c.Usage)
	}

	if cmd.Flags().Changed("description") {
		req.Description = c.Description
	}

	if cmd.Flags().Changed("multizone") {
		req.MultiZone = &c.Multizone
	}

	resp, err := client.CreateKmsKey(ctx, req)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&KeyShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Key:                resp.ID.String(),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
