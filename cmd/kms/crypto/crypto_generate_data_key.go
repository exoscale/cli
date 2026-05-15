package crypto

import (
	"encoding/base64"
	"os"
	"strconv"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type cryptoGenerateDataKeyOutput struct {
	Plaintext  string `json:"plaintext"`
	Ciphertext string `json:"ciphertext"`
}

func (o *cryptoGenerateDataKeyOutput) ToJSON() { output.JSON(o) }
func (o *cryptoGenerateDataKeyOutput) ToText() { output.Text(o) }
func (o *cryptoGenerateDataKeyOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"PLAINTEXT",
		"CIPHERTEXT",
	})

	t.Append([]string{
		o.Plaintext,
		o.Ciphertext,
	})
}

type cryptoGenerateDataKeyCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"generate-data-key"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	KeySpec           v3.GenerateDataKeyRequestKeySpec `cli-short:"s" cli-flag:"key-spec" cli-usage:"key spec for DEK [AES-256]"`
	BytesCount        string                           `cli-short:"b" cli-flag:"bytes-count" cli-usage:"number of bytes for DEK (1 - 1024)"`
	EncryptionContext string                           `cli-short:"e" cli-flag:"encryption-context" cli-usage:"encryption context to use for DEK generation"`
	Zone              v3.ZoneName                      `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *cryptoGenerateDataKeyCmd) CmdAliases() []string { return nil }

func (c *cryptoGenerateDataKeyCmd) CmdShort() string {
	return "Generate a Data Encryption Key from a given KMS Key."
}

func (c *cryptoGenerateDataKeyCmd) CmdLong() string {
	return "Generate a Data Encryption Key from a given KMS Key."
}

func (c *cryptoGenerateDataKeyCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *cryptoGenerateDataKeyCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	ec := []byte(c.EncryptionContext)

	var bytecount int
	if c.BytesCount != "" {
		n, err := strconv.Atoi(c.BytesCount)
		if err != nil {
			return err
		}
		bytecount = n
	}

	req := v3.GenerateDataKeyRequest{
		KeySpec:           c.KeySpec,
		BytesCount:        bytecount,
		EncryptionContext: &ec,
	}

	resp, err := client.GenerateDataKey(ctx, v3.UUID(c.Key), req)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		out := cryptoGenerateDataKeyOutput{
			Ciphertext: base64.StdEncoding.EncodeToString(resp.Ciphertext),
			Plaintext:  base64.StdEncoding.EncodeToString(resp.Plaintext),
		}
		return c.OutputFunc(&out, nil)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(cryptoCmd, &cryptoGenerateDataKeyCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
