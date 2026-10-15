package keystore

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyStoreConnectCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"connect"`

	KeyStore string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *keyStoreConnectCmd) CmdAliases() []string { return nil }
func (c *keyStoreConnectCmd) CmdShort() string     { return "Connect a key store" }
func (c *keyStoreConnectCmd) CmdLong() string {
	return "This command connects an external key store to its XKS proxy, making it usable by KMS keys."
}
func (c *keyStoreConnectCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyStoreConnectCmd) CmdRun(_ *cobra.Command, _ []string) error {
	id, err := parseKeyStoreID(c.KeyStore)
	if err != nil {
		return err
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if _, err := client.ConnectKeyStore(ctx, id); err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&KeyStoreShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			KeyStore:           id.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &keyStoreConnectCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
