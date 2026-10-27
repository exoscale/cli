package crypto

import (
	"encoding/base64"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type cryptoDecryptOutput struct {
	Plaintext string `json:"plaintext"`
}

func (o *cryptoDecryptOutput) ToJSON() { output.JSON(o) }
func (o *cryptoDecryptOutput) ToText() { output.Text(o) }
func (o *cryptoDecryptOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"PLAINTEXT",
	})

	t.Append([]string{
		o.Plaintext,
	})
}

type cryptoDecryptCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"decrypt"`

	Key        string             `cli-arg:"#" cli-usage:"ID"`
	Ciphertext exocmd.BinaryValue `cli-arg:"#" cli-usage:"CIPHERTEXT"`

	EncryptionContext   exocmd.BinaryValue                   `cli-short:"e" cli-flag:"encryption-context" cli-usage:"encryption context to use for decryption"`
	EncryptionAlgorithm v3.DecryptRequestEncryptionAlgorithm `cli-flag:"encryption-algorithm" cli-usage:"encryption algorithm, required for asymmetric keys [AES_256|RSAES_OAEP_SHA_256] (default: AES_256)"`
	Zone                v3.ZoneName                          `cli-short:"z" cli-flag:"zone" cli-usage:"crypto zone"`
}

func (c *cryptoDecryptCmd) CmdAliases() []string { return nil }

func (c *cryptoDecryptCmd) CmdShort() string {
	return "Decrypt a ciphertext."
}

func (c *cryptoDecryptCmd) CmdLong() string {
	return "Decrypt a ciphertext.\n\n" + binaryInputHelp
}

func (c *cryptoDecryptCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *cryptoDecryptCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	req := v3.DecryptRequest{
		Ciphertext:          c.Ciphertext,
		EncryptionAlgorithm: c.EncryptionAlgorithm,
	}

	if cmd.Flags().Changed("encryption-context") {
		ec := []byte(c.EncryptionContext)
		req.EncryptionContext = &ec
	}

	resp, err := client.Decrypt(ctx, v3.UUID(c.Key), req)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		out := cryptoDecryptOutput{
			Plaintext: base64.StdEncoding.EncodeToString(resp.Plaintext),
		}
		return c.OutputFunc(&out, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(cryptoCmd, &cryptoDecryptCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
