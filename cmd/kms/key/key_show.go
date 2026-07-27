package key

import (
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type KeyShowOutput struct {
	ID             v3.UUID                    `json:"id" validate:"required"`
	Name           string                     `json:"name" validate:"required"`
	CreatedAt      time.Time                  `json:"created-at" validate:"required"`
	Multizone      bool                       `json:"multi-zone" validate:"required"`
	OriginZone     string                     `json:"origin-zone" validate:"required"`
	Status         v3.GetKmsKeyResponseStatus `json:"status" validate:"required"`
	ReplicasStatus string                     `json:"replicas-status,omitempty"`
	Material       string                     `json:"material" validate:"required"`
	Rotation       string                     `json:"rotation" validate:"required"`
	Usage          string                     `json:"usage" validate:"required"`
	Source         v3.GetKmsKeyResponseSource `json:"source" validate:"required"`
	Description    string                     `json:"description,omitempty"`
	DeleteAt       time.Time                  `json:"delete-at,omitempty"`
}

func (o *KeyShowOutput) Type() string { return "KMS key" }
func (o *KeyShowOutput) ToJSON()      { output.JSON(o) }
func (o *KeyShowOutput) ToText()      { output.Text(o) }
func (o *KeyShowOutput) ToTable()     { output.Table(o) }

type KeyShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Key string `cli-arg:"#" cli-usage:"ID"`

	Zone v3.ZoneName `cli-short:"z" cli-flag:"zone" cli-usage:"key zone"`
}

func (c *KeyShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *KeyShowCmd) CmdShort() string {
	return "Retrieve KMS Key details."
}

func (c *KeyShowCmd) CmdLong() string {
	return "Retrieve KMS Key details."
}

func (c *KeyShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *KeyShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.GetKmsKey(ctx, v3.UUID(c.Key))
	if err != nil {
		return err
	}

	out := KeyShowOutput{
		ID:             resp.ID,
		Name:           resp.Name,
		CreatedAt:      resp.CreatedAT,
		Multizone:      *resp.MultiZone,
		OriginZone:     resp.OriginZone,
		Status:         resp.Status,
		ReplicasStatus: formatReplicaStatus(resp.ReplicasStatus),
		Material:       formatKeyMaterial(resp.Material),
		Rotation:       formatKeyRotationConfig(resp.Rotation),
		Usage:          resp.Usage,
		Source:         resp.Source,
		Description:    resp.Description,
		DeleteAt:       resp.DeleteAT,
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(keyCmd, &KeyShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
