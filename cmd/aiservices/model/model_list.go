package model

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type ModelListItemOutput struct {
	ID        v3.UUID                         `json:"id" outputWidth:"36"`
	Name      string                          `json:"name" outputWidth:"70"`
	Zone      v3.ZoneName                     `json:"zone" outputWidth:"8"`
	State     v3.ListModelsResponseEntryState `json:"state" outputLabel:"Status" outputWidth:"14"`
	ModelSize string                          `json:"model_size" outputLabel:"Size" outputWidth:"12"`
}

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
		strings.Join(output.TemplateAnnotations(&ModelListItemOutput{}), ", "))
}
func (c *ModelListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *ModelListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runModelList(c, os.Stdout, os.Stderr)
}

func runModelList(c *ModelListCmd, stdout, stderr io.Writer) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(ModelListItemOutput{}, stdout)
	defer func() { _ = streamer.Close() }()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			resp, err := zc.ListModels(ctx)
			if err != nil {
				return err
			}
			for _, m := range resp.Models {
				var size string
				if m.ModelSize != 0 {
					size = humanize.IBytes(uint64(m.ModelSize))
				}
				if err := streamer.Push(ModelListItemOutput{
					ID:        m.ID,
					Name:      m.Name,
					Zone:      zone.Name,
					State:     m.State,
					ModelSize: size,
				}); err != nil {
					return err
				}
			}
			return nil
		})

	if failed > 0 {
		return fmt.Errorf("%d zone(s) failed", failed)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &ModelListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
