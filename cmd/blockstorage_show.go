package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageShowOutput struct {
	BlockStorageSnapshots []v3.BlockStorageSnapshotTarget `json:"block-storage-snapshots"`
	Blocksize             int64                           `json:"blocksize"`
	CreatedAT             time.Time                       `json:"created-at"`
	ID                    v3.UUID                         `json:"id"`
	Instance              *v3.InstanceTarget              `json:"instance"`
	Labels                v3.Labels                       `json:"labels"`
	Name                  string                          `json:"name"`
	Size                  string                          `json:"size"`
	State                 v3.BlockStorageVolumeState      `json:"state"`
}

func (o *blockstorageShowOutput) Type() string { return "Block Storage Volume" }
func (o *blockstorageShowOutput) ToJSON()      { output.JSON(o) }
func (o *blockstorageShowOutput) ToText()      { output.Text(o) }
func (o *blockstorageShowOutput) ToTable()     { output.Table(o) }

type blockstorageShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"block storage volume zone"`
}

func (c *blockstorageShowCmd) cmdAliases() []string { return gShowAlias }

func (c *blockstorageShowCmd) cmdShort() string { return "Show a Block Storage Volume details" }

func (c *blockstorageShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Block Storage Volume details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *blockstorageShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageShowCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	volumes, err := client.ListBlockStorageVolumes(ctx)
	if err != nil {
		return err
	}
	volume, err := volumes.FindBlockStorageVolume(c.Name)
	if err != nil {
		return err
	}

	return c.outputFunc(&blockstorageShowOutput{
		ID:                    volume.ID,
		Name:                  volume.Name,
		Size:                  humanize.IBytes(uint64(volume.Size)),
		Blocksize:             volume.Blocksize,
		CreatedAT:             volume.CreatedAT,
		State:                 volume.State,
		Instance:              volume.Instance,
		Labels:                volume.Labels,
		BlockStorageSnapshots: volume.BlockStorageSnapshots,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
