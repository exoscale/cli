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

type blockstorageSnapshotShowOutput struct {
	CreatedAT time.Time                    `json:"created-at"`
	ID        v3.UUID                      `json:"id"`
	Labels    v3.Labels                    `json:"labels"`
	Volume    *v3.BlockStorageVolumeTarget `json:"volume"`
	Name      string                       `json:"name"`
	Size      string                       `json:"size"`
	State     v3.BlockStorageSnapshotState `json:"state"`
}

func (o *blockstorageSnapshotShowOutput) Type() string { return "Block Storage Volume Snapshot" }
func (o *blockstorageSnapshotShowOutput) ToJSON()      { output.JSON(o) }
func (o *blockstorageSnapshotShowOutput) ToText()      { output.Text(o) }
func (o *blockstorageSnapshotShowOutput) ToTable()     { output.Table(o) }

type blockstorageSnapshotShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"block storage zone"`
}

func (c *blockstorageSnapshotShowCmd) cmdAliases() []string { return gShowAlias }

func (c *blockstorageSnapshotShowCmd) cmdShort() string { return "Show a Block Storage Volume details" }

func (c *blockstorageSnapshotShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Block Storage Volume Snapshot details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *blockstorageSnapshotShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageSnapshotShowCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	snapshots, err := client.ListBlockStorageSnapshots(ctx)
	if err != nil {
		return err
	}
	snapshot, err := snapshots.FindBlockStorageSnapshot(c.Name)
	if err != nil {
		return err
	}

	return c.outputFunc(&blockstorageSnapshotShowOutput{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Size:      humanize.IBytes(uint64(snapshot.Size)),
		CreatedAT: snapshot.CreatedAT,
		State:     snapshot.State,
		Labels:    snapshot.Labels,
		Volume:    snapshot.BlockStorageVolume,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageSnapshotCmd, &blockstorageSnapshotShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
