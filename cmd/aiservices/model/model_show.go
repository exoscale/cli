package model

import (
	"time"

	"github.com/dustin/go-humanize"
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type ModelShowOutput struct {
	ID        v3.UUID                  `json:"id"`
	Name      string                   `json:"name"`
	State     v3.GetModelResponseState `json:"state" outputLabel:"Status"`
	ModelSize string                   `json:"model_size" outputLabel:"Size"`
	CreatedAt string                   `json:"created_at"`
	UpdatedAt string                   `json:"updated_at"`
}

func (o *ModelShowOutput) ToJSON()  { output.JSON(o) }
func (o *ModelShowOutput) ToText()  { output.Text(o) }
func (o *ModelShowOutput) ToTable() { output.Table(o) }

type ModelShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Model string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *ModelShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *ModelShowCmd) CmdShort() string     { return "Show AI model" }
func (c *ModelShowCmd) CmdLong() string {
	return "This command shows details of an AI model by ID or name."
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

	// Resolve model ID using the SDK helper
	list, err := client.ListModels(ctx)
	if err != nil {
		return err
	}
	entry, err := list.FindListModelsResponseEntry(c.Model)
	if err != nil {
		return err
	}
	id := entry.ID

	resp, err := client.GetModel(ctx, id)
	if err != nil {
		return err
	}
	var size string
	if resp.ModelSize != 0 {
		size = humanize.IBytes(uint64(resp.ModelSize))
	}
	out := &ModelShowOutput{
		ID:        resp.ID,
		Name:      resp.Name,
		State:     resp.State,
		ModelSize: size,
		CreatedAt: resp.CreatedAT.Format(time.RFC3339),
		UpdatedAt: resp.UpdatedAT.Format(time.RFC3339),
	}
	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
