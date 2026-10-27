package crypto

import (
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

	Key              string                           `cli-arg:"#" cli-usage:"ID"`
	Message          exocmd.BinaryValue               `cli-arg:"#" cli-usage:"MESSAGE"`
	Signature        exocmd.BinaryValue               `cli-arg:"#" cli-usage:"SIGNATURE"`
	SigningAlgorithm v3.VerifyRequestSigningAlgorithm `cli-arg:"#" cli-usage:"SIGNING_ALGORITHM"`

	MessageType v3.VerifyRequestMessageType `cli-flag:"message-type" cli-usage:"how the message is interpreted [raw|digest] (default: raw)"`
	Zone        v3.ZoneName                 `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *cryptoVerifyCmd) CmdAliases() []string { return nil }

func (c *cryptoVerifyCmd) CmdShort() string {
	return "Verify a signature."
}

func (c *cryptoVerifyCmd) CmdLong() string {
	return `Verify a signature against a message or digest with a KMS Key of usage "sign-verify". The signing algorithm must be the one the signature was produced with.

Supported signing algorithms: RSASSA_PSS_SHA_256, RSASSA_PSS_SHA_384, RSASSA_PSS_SHA_512, RSASSA_PKCS1_V1_5_SHA_256, RSASSA_PKCS1_V1_5_SHA_384, RSASSA_PKCS1_V1_5_SHA_512, ECDSA_SHA_256, ECDSA_SHA_384, ECDSA_SHA_512, EDDSA_ED25519, ED25519_PH_SHA_512, ML_DSA_SHAKE_256

` + binaryInputHelp
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

	resp, err := client.Verify(ctx, v3.UUID(c.Key), v3.VerifyRequest{
		Message:          c.Message,
		MessageType:      c.MessageType,
		Signature:        c.Signature,
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
