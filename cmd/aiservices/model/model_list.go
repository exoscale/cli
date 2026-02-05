package model

import (
	"fmt"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type ModelListItemOutput struct {
	ID        v3.UUID                          `json:"id"`
	Name      string                           `json:"name"`
	Zone      v3.ZoneName                      `json:"zone"`
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
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *ModelListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	out := make(ModelListOutput, 0)
	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {
		c := client.WithEndpoint(zone.APIEndpoint)
		resp, err := c.ListModels(ctx)
		if err != nil {
			return err
		}

		for _, m := range resp.Models {
			var sizePtr *int64
			if m.ModelSize != 0 {
				size := m.ModelSize
				sizePtr = &size
			}
			out = append(out, ModelListItemOutput{
				ID:        m.ID,
				Name:      m.Name,
				Zone:      zone.Name,
				Status:    m.Status,
				ModelSize: sizePtr,
			})
		}

		return nil
	})

	return c.OutputFunc(&out, err)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
