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

type cryptoVerifyOutput struct {
	SignatureValid bool   `json:"signature-valid"`
	KeyID          string `json:"key-id"`
	KeySpec        string `json:"key-spec"`
}

func (o *cryptoVerifyOutput) ToJSON()  { output.JSON(o) }
func (o *cryptoVerifyOutput) ToText()  { output.Text(o) }
func (o *cryptoVerifyOutput) ToTable() { output.Table(o) }

type cryptoVerifyCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"verify"`

	Key       string `cli-arg:"#" cli-usage:"ID"`
	Message   string `cli-arg:"#" cli-usage:"MESSAGE_b64"`
	Signature string `cli-arg:"#" cli-usage:"SIGNATURE_b64"`

	SigningAlgorithm v3.VerifyRequestSigningAlgorithm `cli-flag:"signing-algorithm" cli-usage:"signing algorithm the signature was produced with, required [RSASSA_PSS_SHA_256|RSASSA_PSS_SHA_384|RSASSA_PSS_SHA_512|ECDSA_SHA_256|ECDSA_SHA_384|ECDSA_SHA_512|EDDSA_ED25519|ED25519_PH_SHA_512|ML_DSA_SHAKE_256]"`
	MessageType      v3.VerifyRequestMessageType      `cli-flag:"message-type" cli-usage:"how the message is interpreted [raw|digest] (default: raw)"`
	Zone             v3.ZoneName                      `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *cryptoVerifyCmd) CmdAliases() []string { return nil }

func (c *cryptoVerifyCmd) CmdShort() string {
	return "Verify a signature."
}

func (c *cryptoVerifyCmd) CmdLong() string {
	return `Verify a base64-encoded signature against a base64-encoded message or digest with a KMS Key of usage "sign-verify".`
}

func (c *cryptoVerifyCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *cryptoVerifyCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	signature, err := base64.StdEncoding.DecodeString(c.Signature)
	if err != nil {
		return fmt.Errorf("signature is not valid base64: %w", err)
	}

	resp, err := client.Verify(ctx, v3.UUID(c.Key), v3.VerifyRequest{
		Message:          message,
		MessageType:      c.MessageType,
		Signature:        signature,
		SigningAlgorithm: c.SigningAlgorithm,
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		out := cryptoVerifyOutput{
			SignatureValid: resp.SignatureValid != nil && *resp.SignatureValid,
			KeyID:          resp.KeyID.String(),
			KeySpec:        resp.KeySpec,
		}
		return c.OutputFunc(&out, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(cryptoCmd, &cryptoVerifyCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
