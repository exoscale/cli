package model

import (
	"fmt"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type ModelShowOutput struct {
	ID        v3.UUID                   `json:"id"`
	Name      string                    `json:"name"`
	Status    v3.GetModelResponseStatus `json:"status"`
	ModelSize *int64                    `json:"model_size"`
	CreatedAt string                    `json:"created_at"`
	UpdatedAt string                    `json:"updated_at"`
}

func (o *ModelShowOutput) ToJSON()  { output.JSON(o) }
func (o *ModelShowOutput) ToText()  { output.Text(o) }
func (o *ModelShowOutput) ToTable() { output.Table(o) }

type ModelShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ID   string      `cli-arg:"#" cli-usage:"MODEL-ID (UUID)"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *ModelShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *ModelShowCmd) CmdShort() string     { return "Show AI model" }
func (c *ModelShowCmd) CmdLong() string {
	return "This command shows details of an AI model by its ID."
}
func (c *ModelShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *ModelShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid model ID: %w", err)
	}
	resp, err := client.GetModel(ctx, id)
	if err != nil {
		return err
	}
	sizePtr := int64PtrIfNonZero(resp.ModelSize)
	out := &ModelShowOutput{
		ID:        resp.ID,
		Name:      resp.Name,
		Status:    resp.Status,
		ModelSize: sizePtr,
		CreatedAt: resp.CreatedAT.Format(time.RFC3339),
		UpdatedAt: resp.UpdatedAT.Format(time.RFC3339),
	}
	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
