package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageShowOutput struct {
	BlockStorageSnapshots []string          `json:"block-storage-snapshots"`
	Blocksize             int64             `json:"blocksize"`
	CreatedAT             time.Time         `json:"created-at"`
	ID                    string            `json:"id"`
	Instance              string            `json:"instance"`
	Labels                map[string]string `json:"labels"`
	Name                  string            `json:"name"`
	Size                  int64             `json:"size"`
	State                 string            `json:"state"`
}

func (o *blockstorageShowOutput) Type() string { return "Block Storage Volume" }
func (o *blockstorageShowOutput) ToJSON()      { output.JSON(o) }
func (o *blockstorageShowOutput) ToText()      { output.Text(o) }
func (o *blockstorageShowOutput) ToTable()     { output.Table(o) }

type blockstorageShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone string `cli-short:"z" cli-usage:"block storage volume zone"`
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
	client := globalstate.EgoscaleV3Client
	TODO := context.TODO()

	bs, err := client.GetBlockStorageVolume(TODO, v3.UUID(c.Name))
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}

		return err
	}

	return c.outputFunc(&blockstorageShowOutput{
		ID:        bs.ID.String(),
		Name:      bs.Name,
		Size:      bs.Size,
		Blocksize: bs.Blocksize,
		CreatedAT: bs.CreatedAT,
		State:     string(bs.State),
		Instance: func(i *v3.InstanceTarget) string {
			if i != nil {
				return i.ID.String()
			}
			return ""
		}(bs.Instance),
		Labels: bs.Labels,
		BlockStorageSnapshots: func(snapshots []v3.BlockStorageSnapshotTarget) []string {
			var v []string
			for _, s := range snapshots {
				v = append(v, s.ID.String())
			}
			return v
		}(bs.BlockStorageSnapshots),
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
