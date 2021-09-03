package cmd

import (
	"fmt"
	"strings"

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
func (o *instanceSnapshotShowOutput) toJSON()      { outputJSON(o) }
func (o *instanceSnapshotShowOutput) toText()      { outputText(o) }
func (o *instanceSnapshotShowOutput) toTable()     { outputTable(o) }

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
		strings.Join(outputterTemplateAnnotations(&instanceSnapshotShowOutput{}), ", "))
}

func (c *instanceSnapshotShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return output(showInstanceSnapshot(c.Zone, c.ID))
}

func showInstanceSnapshot(zone, snapshotID string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	snapshot, err := cs.GetSnapshot(ctx, zone, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Compute instance snapshot: %s", err)
	}

	instance, err := cs.GetInstance(ctx, zone, *snapshot.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Compute instance %q: %s", *snapshot.InstanceID, err)
	}

	return &instanceSnapshotShowOutput{
		ID:           *snapshot.ID,
		Name:         *snapshot.Name,
		CreationDate: snapshot.CreatedAt.String(),
		State:        *snapshot.State,
		Size:         *snapshot.Size,
		Instance:     *instance.Name,
		Zone:         zone,
	}, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceSnapshotCmd, &instanceSnapshotShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
