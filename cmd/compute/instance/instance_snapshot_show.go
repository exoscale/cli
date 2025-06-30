package instance

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSnapshotShowOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
	State        string `json:"state"`
	Size         int64  `json:"size" outputLabel:"Size (GB)"`
	Instance     string `json:"instance"`
	Zone         string `json:"zone"`
}

func (o *instanceSnapshotShowOutput) Type() string { return "Snapshot" }
func (o *instanceSnapshotShowOutput) ToJSON()      { output.JSON(o) }
func (o *instanceSnapshotShowOutput) ToText()      { output.Text(o) }
func (o *instanceSnapshotShowOutput) ToTable()     { output.Table(o) }

type instanceSnapshotShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ID string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *instanceSnapshotShowCmd) CmdShort() string {
	return "Show a Compute instance snapshot details"
}

func (c *instanceSnapshotShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance snapshot details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceSnapshotShowOutput{}), ", "))
}

func (c *instanceSnapshotShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return err
	}
	snapshot, err := snapshots.FindSnapshot(c.ID)
	if err != nil {
		return err
	}
	instance, err := client.GetInstance(ctx, snapshot.Instance.ID)
	if err != nil {
		return fmt.Errorf("unable to retrieve Compute instance %s: %w", snapshot.Instance.ID, err)
	}

	return c.OutputFunc(&instanceSnapshotShowOutput{
		ID:           snapshot.ID.String(),
		Name:         snapshot.Name,
		CreationDate: snapshot.CreatedAT.String(),
		State:        string(snapshot.State),
		Size:         snapshot.Size,
		Instance:     instance.Name,
		Zone:         c.Zone,
	}, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
