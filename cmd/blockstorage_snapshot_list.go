package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageSnapshotListItemOutput struct {
	ID     v3.UUID     `json:"id"`
	Name   string      `json:"name"`
	Zone   v3.ZoneName `json:"zone"`
	Volume v3.UUID     `json:"volume"`
}

type blockstorageSnapshotListOutput []blockstorageSnapshotListItemOutput

func (o *blockstorageSnapshotListOutput) ToJSON()  { output.JSON(o) }
func (o *blockstorageSnapshotListOutput) ToText()  { output.Text(o) }
func (o *blockstorageSnapshotListOutput) ToTable() { output.Table(o) }

type blockstorageSnapshotListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *blockstorageSnapshotListCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageSnapshotListCmd) cmdShort() string { return "List Block Storage Volume Snapshots" }

func (c *blockstorageSnapshotListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Block Storage Volume Snapshots.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageListOutput{}), ", "))
}

func (c *blockstorageSnapshotListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageSnapshotListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := gContext

	resp, err := client.ListZones(ctx)
	if err != nil {
		return err
	}
	zones := resp.Zones

	if c.Zone != "" {
		endpoint, err := client.GetZoneAPIEndpoint(ctx, c.Zone)
		if err != nil {
			return err
		}

		zones = []v3.Zone{{APIEndpoint: endpoint}}
	}

	output := make(blockstorageSnapshotListOutput, 0)
	for _, zone := range zones {
		c := client.WithEndpoint(zone.APIEndpoint)

		resp, err := c.ListBlockStorageSnapshots(ctx)
		if err != nil {
			// TODO: remove it once Block Storage is deployed in every zone.
			if strings.Contains(err.Error(), "Availability of the block storage volumes") {
				continue
			}

			return err
		}

		for _, volume := range resp.BlockStorageSnapshots {
			output = append(output, blockstorageSnapshotListItemOutput{
				ID:     volume.ID,
				Name:   volume.Name,
				Zone:   zone.Name,
				Volume: volume.BlockStorageVolume.ID,
			})
		}
	}

	return c.outputFunc(&output, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageSnapshotCmd, &blockstorageSnapshotListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
