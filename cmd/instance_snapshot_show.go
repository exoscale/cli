package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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
func (o *instanceSnapshotShowOutput) toJSON()      { output.JSON(o) }
func (o *instanceSnapshotShowOutput) toText()      { output.Text(o) }
func (o *instanceSnapshotShowOutput) toTable()     { output.Table(o) }

type instanceSnapshotShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ID string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotShowCmd) cmdAliases() []string { return gShowAlias }

func (c *instanceSnapshotShowCmd) cmdShort() string {
	return "Show a Compute instance snapshot details"
}

func (c *instanceSnapshotShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance snapshot details.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceSnapshotShowOutput{}), ", "))
}

func (c *instanceSnapshotShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	snapshot, err := cs.GetSnapshot(ctx, c.Zone, c.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return fmt.Errorf("error retrieving Compute instance snapshot: %w", err)
	}

	instance, err := cs.GetInstance(ctx, c.Zone, *snapshot.InstanceID)
	if err != nil {
		return fmt.Errorf("unable to retrieve Compute instance %s: %w", *snapshot.InstanceID, err)
	}

	return c.outputFunc(&instanceSnapshotShowOutput{
		ID:           *snapshot.ID,
		Name:         *snapshot.Name,
		CreationDate: snapshot.CreatedAt.String(),
		State:        *snapshot.State,
		Size:         *snapshot.Size,
		Instance:     *instance.Name,
		Zone:         c.Zone,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceSnapshotCmd, &instanceSnapshotShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
