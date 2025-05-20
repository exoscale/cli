package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageShowOutput struct {
	BlockStorageSnapshots []v3.BlockStorageSnapshotTarget `json:"block-storage-snapshots"`
	Blocksize             int64                           `json:"blocksize"`
	CreatedAT             time.Time                       `json:"created-at"`
	ID                    v3.UUID                         `json:"id"`
	Instance              *v3.InstanceTarget              `json:"instance"`
	Name                  string                          `json:"name"`
	Size                  string                          `json:"size"`
	Labels                map[string]string               `json:"labels"`
	State                 v3.BlockStorageVolumeState      `json:"state"`
}

func (o *blockStorageShowOutput) Type() string { return "Block Storage Volume" }
func (o *blockStorageShowOutput) ToJSON()      { output.JSON(o) }
func (o *blockStorageShowOutput) ToText()      { output.Text(o) }
func (o *blockStorageShowOutput) ToTable()     { output.Table(o) }

type blockStorageShowCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"block storage volume zone"`
}

func (c *blockStorageShowCmd) CmdAliases() []string { return GShowAlias }

func (c *blockStorageShowCmd) CmdShort() string { return "Show a Block Storage Volume details" }

func (c *blockStorageShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Block Storage Volume details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *blockStorageShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageShowCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	volumes, err := client.ListBlockStorageVolumes(ctx)
	if err != nil {
		return err
	}
	v, err := volumes.FindBlockStorageVolume(c.Name)
	if err != nil {
		return err
	}
	volume, err := client.GetBlockStorageVolume(ctx, v.ID)
	if err != nil {
		return err
	}

	return c.OutputFunc(&blockStorageShowOutput{
		ID:                    volume.ID,
		Name:                  volume.Name,
		Size:                  fmt.Sprintf("%d GiB", volume.Size),
		Blocksize:             volume.Blocksize,
		CreatedAT:             volume.CreatedAT,
		State:                 volume.State,
		Instance:              volume.Instance,
		BlockStorageSnapshots: volume.BlockStorageSnapshots,
		Labels:                volume.Labels,
	}, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(blockstorageCmd, &blockStorageShowCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
