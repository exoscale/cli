package deployment

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type InstanceTypeListItemOutput struct {
	Family     string `json:"family" outputWidth:"16"`
	Authorized bool   `json:"authorized" outputWidth:"10"`
	Zone       string `json:"zone" outputWidth:"8"`
}

type InstanceTypeListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"instance-type"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *InstanceTypeListCmd) CmdAliases() []string { return nil }
func (c *InstanceTypeListCmd) CmdShort() string     { return "List AI instance types" }
func (c *InstanceTypeListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI instance types.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceTypeListItemOutput{}), ", "))
}
func (c *InstanceTypeListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *InstanceTypeListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runInstanceTypeList(c, os.Stdout, os.Stderr)
}

func runInstanceTypeList(c *InstanceTypeListCmd, stdout, stderr io.Writer) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	defer sink.Flush()

	streamer := output.NewStreamer(InstanceTypeListItemOutput{}, stdout)
	defer func() {
		if err := streamer.Close(); err != nil {
			_, _ = fmt.Fprintf(stderr, "error: %s\n", err)
		}
	}()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink, true,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			resp, err := zc.ListAIInstanceTypes(ctx)
			if err != nil {
				return err
			}
			for _, it := range resp.InstanceTypes {
				authorized := false
				if it.Authorized != nil {
					authorized = *it.Authorized
				}
				if err := streamer.Push(InstanceTypeListItemOutput{
					Family:     it.Family,
					Authorized: authorized,
					Zone:       string(zone.Name),
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
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &InstanceTypeListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
