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

type DeploymentListItemOutput struct {
	ID        v3.UUID                              `json:"id"`
	Name      string                               `json:"name"`
	Zone      v3.ZoneName                          `json:"zone"`
	State     v3.ListDeploymentsResponseEntryState `json:"state" outputLabel:"Status"`
	GPUType   string                               `json:"gpu_type"`
	GPUCount  int64                                `json:"gpu_count"`
	Replicas  int64                                `json:"replicas"`
	ModelName string                               `json:"model_name"`
}

type DeploymentListOutput []DeploymentListItemOutput

func (o *DeploymentListOutput) ToJSON()  { output.JSON(o) }
func (o *DeploymentListOutput) ToText()  { output.Text(o) }
func (o *DeploymentListOutput) ToTable() { output.Table(o) }

type DeploymentListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *DeploymentListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *DeploymentListCmd) CmdShort() string     { return "List AI deployments" }
func (c *DeploymentListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI deployments.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&DeploymentListItemOutput{}), ", "))
}
func (c *DeploymentListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *DeploymentListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return runDeploymentList(c, os.Stdout, os.Stderr)
}

func runDeploymentList(c *DeploymentListCmd, stdout, stderr io.Writer) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	sink := utils.NewWarningSinkTo(stderr)
	stopSig := sink.InstallSignalFlush(ctx)
	defer stopSig()
	defer sink.Flush()

	streamer := output.NewStreamer(DeploymentListItemOutput{}, stdout)
	defer func() { _ = streamer.Close() }()

	failed := utils.ForEveryZoneAsync(ctx, zones, globalstate.RequestTimeout, sink,
		func(ctx context.Context, zone v3.Zone) error {
			zc := client.WithEndpoint(zone.APIEndpoint)
			resp, err := zc.ListDeployments(ctx)
			if err != nil {
				return err
			}
			for _, d := range resp.Deployments {
				var modelName string
				if d.Model != nil {
					modelName = d.Model.Name
				}
				if err := streamer.Push(DeploymentListItemOutput{
					ID:        d.ID,
					Name:      d.Name,
					Zone:      zone.Name,
					State:     d.State,
					GPUType:   d.GpuType,
					GPUCount:  d.GpuCount,
					Replicas:  d.Replicas,
					ModelName: modelName,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &DeploymentListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
