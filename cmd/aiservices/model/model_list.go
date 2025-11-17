package model

import (
	"fmt"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type ModelListItemOutput struct {
	ID        v3.UUID                          `json:"id"`
	Name      string                           `json:"name"`
	Status    v3.ListModelsResponseEntryStatus `json:"status"`
	ModelSize *int64                           `json:"model_size"`
}

type ModelListOutput []ModelListItemOutput

func (o *ModelListOutput) ToJSON()  { output.JSON(o) }
func (o *ModelListOutput) ToText()  { output.Text(o) }
func (o *ModelListOutput) ToTable() { output.Table(o) }

type ModelListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *ModelListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *ModelListCmd) CmdShort() string     { return "List AI models" }
func (c *ModelListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI models.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&ModelListOutput{}), ", "))
}
func (c *ModelListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *ModelListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListModels(ctx)
	if err != nil {
		return err
	}

	out := make(ModelListOutput, 0, len(resp.Models))
	for _, m := range resp.Models {
		sizePtr := int64PtrIfNonZero(m.ModelSize)
		out = append(out, ModelListItemOutput{
			ID:        m.ID,
			Name:      m.Name,
			Status:    m.Status,
			ModelSize: sizePtr,
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
