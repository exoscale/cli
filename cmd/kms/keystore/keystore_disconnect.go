package keystore

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyStoreDisconnectCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"disconnect"`

	KeyStore string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *keyStoreDisconnectCmd) CmdAliases() []string { return nil }
func (c *keyStoreDisconnectCmd) CmdShort() string     { return "Disconnect a key store" }
func (c *keyStoreDisconnectCmd) CmdLong() string {
	return "This command disconnects an external key store from its XKS proxy. KMS keys backed by it are unusable until it is connected again."
}
func (c *keyStoreDisconnectCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyStoreDisconnectCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := resolveKeyStoreID(ctx, client, c.KeyStore)
	if err != nil {
		return err
	}

	if _, err := client.DisconnectKeyStore(ctx, id); err != nil {
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
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &keyStoreDisconnectCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
