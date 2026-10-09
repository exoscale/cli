package crypto

import (
	"encoding/base64"
	"errors"
	"fmt"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type cryptoSignOutput struct {
	Signature        string `json:"signature"`
	KeySpec          string `json:"key-spec"`
	SigningAlgorithm string `json:"signing-algorithm"`
}

func (o *cryptoSignOutput) ToJSON()  { output.JSON(o) }
func (o *cryptoSignOutput) ToText()  { output.Text(o) }
func (o *cryptoSignOutput) ToTable() { output.Table(o) }

type cryptoSignCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"sign"`

	Key     string `cli-arg:"#" cli-usage:"ID"`
	Message string `cli-arg:"#" cli-usage:"MESSAGE_b64"`

	SigningAlgorithm v3.SignRequestSigningAlgorithm `cli-flag:"signing-algorithm" cli-usage:"signing algorithm, required [RSASSA_PSS_SHA_256|RSASSA_PSS_SHA_384|RSASSA_PSS_SHA_512|ECDSA_SHA_256|ECDSA_SHA_384|ECDSA_SHA_512|EDDSA_ED25519|ED25519_PH_SHA_512|ML_DSA_SHAKE_256]"`
	MessageType      v3.SignRequestMessageType      `cli-flag:"message-type" cli-usage:"how the message is interpreted [raw|digest] (default: raw)"`
	Zone             v3.ZoneName                    `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *cryptoSignCmd) CmdAliases() []string { return nil }

func (c *cryptoSignCmd) CmdShort() string {
	return "Sign a message or digest."
}

func (c *cryptoSignCmd) CmdLong() string {
	return `Sign a base64-encoded message or digest with a KMS Key of usage "sign-verify". The signing algorithm must match the key spec of the KMS Key.`
}

func (c *cryptoSignCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *cryptoSignCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.SigningAlgorithm == "" {
		return errors.New("--signing-algorithm is required")
	}

	message, err := base64.StdEncoding.DecodeString(c.Message)
	if err != nil {
		return fmt.Errorf("message is not valid base64: %w", err)
	}

	resp, err := client.Sign(ctx, v3.UUID(c.Key), v3.SignRequest{
		Message:          message,
		MessageType:      c.MessageType,
		SigningAlgorithm: c.SigningAlgorithm,
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		out := cryptoSignOutput{
			Signature:        base64.StdEncoding.EncodeToString(resp.Signature),
			KeySpec:          resp.KeySpec,
			SigningAlgorithm: resp.SigningAlgorithm,
		}
		return c.OutputFunc(&out, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(cryptoCmd, &cryptoSignCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
