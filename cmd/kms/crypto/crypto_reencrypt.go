package crypto

import (
	"encoding/base64"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type cryptoReencryptCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reencrypt"`

	Key            string `cli-arg:"#" cli-usage:"SRC_ID"`
	DestinationKey string `cli-arg:"#" cli-usage:"DEST_ID"`
	Ciphertext     string `cli-arg:"#" cli-usage:"CIPHERTEXT"`

	SourceEncryptionContext string      `cli-short:"s" cli-flag:"source-encryption-context" cli-usage:"encryption context to use for source ciphertext decryption"`
	DestEncryptionContext   string      `cli-short:"d" cli-flag:"dest-encryption-context" cli-usage:"encryption context to use for destination ciphertext encryption"`
	Zone                    v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *cryptoReencryptCmd) CmdAliases() []string { return nil }

func (c *cryptoReencryptCmd) CmdShort() string {
	return "Decrypts and encrypts an exisiting ciphertext with newest key material or a different KMS key."
}

func (c *cryptoReencryptCmd) CmdLong() string {
	return "Decrypts an existing ciphertext using its original key material and re-encrypts the underlying plaintext using a specified KMS key or the latest key material of the same KMS Key."
}

func (c *cryptoReencryptCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *cryptoReencryptCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	decodedCipher, err := base64.StdEncoding.DecodeString(c.Ciphertext)
	if err != nil {
		return err
	}
	source := &v3.ReEncryptRequestSource{
		Ciphertext: decodedCipher,
		Key:        v3.UUID(c.Key),
	}

	if cmd.Flags().Changed("source-encryption-context") {
		ec := []byte(c.SourceEncryptionContext)
		source.EncryptionContext = &ec
	}

	dest := &v3.ReEncryptRequestDestination{
		Key: v3.UUID(c.DestinationKey),
	}

	if cmd.Flags().Changed("dest-encryption-context") {
		ec := []byte(c.DestEncryptionContext)
		dest.EncryptionContext = &ec
	}

	req := v3.ReEncryptRequest{
		Source:      source,
		Destination: dest,
	}

	resp, err := client.ReEncrypt(ctx, v3.UUID(c.Key), req)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		out := cryptoEncryptOutput{
			Ciphertext: base64.StdEncoding.EncodeToString(resp.Ciphertext),
		}
		return c.OutputFunc(&out, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(cryptoCmd, &cryptoReencryptCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
