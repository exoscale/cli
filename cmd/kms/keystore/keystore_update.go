package keystore

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyStoreUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	KeyStore string `cli-arg:"#" cli-usage:"ID"`

	Description string      `cli-short:"d" cli-usage:"key store description"`
	Endpoint    string      `cli-short:"e" cli-usage:"public URL of the XKS proxy"`
	AccessKey   string      `cli-flag:"access-key" cli-usage:"access key used to sign requests sent to the XKS proxy (requires --secret-key)"`
	SecretKey   string      `cli-flag:"secret-key" cli-usage:"secret key used to sign requests sent to the XKS proxy (requires --access-key)"`
	Zone        v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *keyStoreUpdateCmd) CmdAliases() []string { return nil }
func (c *keyStoreUpdateCmd) CmdShort() string     { return "Update a key store" }
func (c *keyStoreUpdateCmd) CmdLong() string {
	return "This command updates an external key store's description, XKS proxy endpoint or XKS proxy credentials."
}
func (c *keyStoreUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyStoreUpdateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	id, err := parseKeyStoreID(c.KeyStore)
	if err != nil {
		return err
	}

	if (c.AccessKey == "") != (c.SecretKey == "") {
		return fmt.Errorf("--access-key and --secret-key must be set together")
	}

	req := v3.UpdateKeyStoreRequest{Description: c.Description}
	if c.Endpoint != "" || c.AccessKey != "" {
		req.Proxy = &v3.UpdateKeyStoreProxy{Endpoint: c.Endpoint}
		if c.AccessKey != "" {
			req.Proxy.Auth = &v3.KeyStoreProxyAuth{
				Key:    c.AccessKey,
				Secret: c.SecretKey,
			}
		}
	}
	if req.Description == "" && req.Proxy == nil {
		return fmt.Errorf("nothing to update: set --description, --endpoint or --access-key/--secret-key")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	ks, err := client.UpdateKeyStore(ctx, id, req)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc(newKeyStoreShowOutput(ks), nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &keyStoreUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
