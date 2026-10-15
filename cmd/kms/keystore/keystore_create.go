package keystore

import (
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyStoreCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name     string `cli-arg:"#" cli-usage:"NAME"`
	Endpoint string `cli-arg:"#" cli-usage:"ENDPOINT"`

	Description     string      `cli-short:"d" cli-usage:"key store description"`
	AccessKey       string      `cli-flag:"access-key" cli-usage:"access key used to sign requests sent to the XKS proxy"`
	SecretKey       string      `cli-flag:"secret-key" cli-usage:"secret key used to sign requests sent to the XKS proxy"`
	ConnectOnCreate bool        `cli-flag:"connect-on-create" cli-usage:"connect the key store right after creating it"`
	Zone            v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *keyStoreCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *keyStoreCreateCmd) CmdShort() string     { return "Create a key store" }
func (c *keyStoreCreateCmd) CmdLong() string {
	return `This command creates an external key store (XKS) backed by a customer-managed XKS proxy.
ENDPOINT is the public URL of the XKS proxy, and --access-key/--secret-key are the
credentials used to sign requests sent to it.

The XKS proxy configuration is validated on creation. A new key store stays
disconnected until "exo kms keystore connect" is run, or --connect-on-create is set.`
}
func (c *keyStoreCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyStoreCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	if c.AccessKey == "" || c.SecretKey == "" {
		return fmt.Errorf("--access-key and --secret-key are required")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.CreateKeyStore(ctx, v3.CreateKeyStoreRequest{
		Name:        c.Name,
		Description: c.Description,
		Type:        v3.CreateKeyStoreRequestTypeExternalKeyStore,
		Proxy: &v3.KeyStoreProxy{
			Endpoint: c.Endpoint,
			Auth: &v3.KeyStoreProxyAuth{
				Key:    c.AccessKey,
				Secret: c.SecretKey,
			},
		},
	})
	if err != nil {
		return err
	}

	if c.ConnectOnCreate {
		if _, err := client.ConnectKeyStore(ctx, resp.ID); err != nil {
			return fmt.Errorf("key store %s created but failed to connect: %w", resp.ID, err)
		}
	}

	if !globalstate.Quiet {
		return (&KeyStoreShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			KeyStore:           resp.ID.String(),
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &keyStoreCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
