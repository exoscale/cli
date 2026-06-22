package crypto

import (
	"encoding/base64"
	"fmt"
	"os"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type cryptoEncryptOutput struct {
	Ciphertext string `json:"ciphertext"`
}

func (o *cryptoEncryptOutput) ToJSON() { output.JSON(o) }
func (o *cryptoEncryptOutput) ToText() { output.Text(o) }
func (o *cryptoEncryptOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()

	t.SetHeader([]string{
		"CIPHERTEXT",
	})

	t.Append([]string{
		o.Ciphertext,
	})
}

type cryptoEncryptCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"encrypt"`

	Key       string `cli-arg:"#" cli-usage:"ID"`
	Plaintext string `cli-arg:"#" cli-usage:"PLAINTEXT_b64"`

	EncryptionContext string      `cli-short:"e" cli-flag:"encryption-context" cli-usage:"encryption context to use for encryption"`
	Zone              v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *cryptoEncryptCmd) CmdAliases() []string { return nil }

func (c *cryptoEncryptCmd) CmdShort() string {
	return "Encrypt a plaintext."
}

func (c *cryptoEncryptCmd) CmdLong() string {
	return "Encrypt a plaintext."
}

func (c *cryptoEncryptCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *cryptoEncryptCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(c.Plaintext)
	if err != nil {
		return fmt.Errorf("plaintext is not valid base64: %w", err)
	}

	req := v3.EncryptRequest{
		Plaintext: decoded,
	}

	if cmd.Flags().Changed("encryption-context") {
		ec, err := base64.StdEncoding.DecodeString(c.EncryptionContext)
		if err != nil {
			return fmt.Errorf("encryption-context is not valid base64: %w", err)
		}
		req.EncryptionContext = &ec
	}

	resp, err := client.Encrypt(ctx, v3.UUID(c.Key), req)
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
	cobra.CheckErr(exocmd.RegisterCLICommand(cryptoCmd, &cryptoEncryptCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
