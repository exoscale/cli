package key

import (
	"encoding/base64"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type keyGetPublicKeyOutput struct {
	KeyID     v3.UUID `json:"key-id"`
	KeySpec   string  `json:"key-spec"`
	Usage     string  `json:"usage"`
	PublicKey string  `json:"public-key"`
}

func (o *keyGetPublicKeyOutput) Type() string { return "KMS public key" }
func (o *keyGetPublicKeyOutput) ToJSON()      { output.JSON(o) }
func (o *keyGetPublicKeyOutput) ToText()      { output.Text(o) }
func (o *keyGetPublicKeyOutput) ToTable()     { output.Table(o) }

type keyGetPublicKeyCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"get-public-key"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *keyGetPublicKeyCmd) CmdAliases() []string { return nil }

func (c *keyGetPublicKeyCmd) CmdShort() string {
	return "Retrieve the public key of an asymmetric KMS Key."
}

func (c *keyGetPublicKeyCmd) CmdLong() string {
	return `Retrieve the public key of an asymmetric KMS Key. The public key is the base64-encoded X.509 SubjectPublicKeyInfo (SPKI) DER encoding.`
}

func (c *keyGetPublicKeyCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *keyGetPublicKeyCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.GetPublicKey(ctx, v3.UUID(c.Key))
	if err != nil {
		return err
	}

	out := keyGetPublicKeyOutput{
		KeyID:     resp.KeyID,
		KeySpec:   resp.KeySpec,
		Usage:     resp.Usage,
		PublicKey: base64.StdEncoding.EncodeToString(resp.PublicKey),
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &keyGetPublicKeyCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
