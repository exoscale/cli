package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageSnapshotListItemOutput struct {
	ID     v3.UUID     `json:"id"`
	Name   string      `json:"name"`
	Zone   v3.ZoneName `json:"zone"`
	Volume v3.UUID     `json:"volume"`
}

type blockStorageSnapshotListOutput []blockStorageSnapshotListItemOutput

func (o *blockStorageSnapshotListOutput) ToJSON()  { output.JSON(o) }
func (o *blockStorageSnapshotListOutput) ToText()  { output.Text(o) }
func (o *blockStorageSnapshotListOutput) ToTable() { output.Table(o) }

type blockStorageSnapshotListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *blockStorageSnapshotListCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockStorageSnapshotListCmd) cmdShort() string { return "List Block Storage Volume Snapshots" }

func (c *blockStorageSnapshotListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Block Storage Volume Snapshots.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageListOutput{}), ", "))
}

func (c *blockStorageSnapshotListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageSnapshotListCmd) cmdRun(_ *cobra.Command, _ []string) error {
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

	output := make(blockStorageSnapshotListOutput, 0)
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
			output = append(output, blockStorageSnapshotListItemOutput{
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
	cobra.CheckErr(registerCLICommand(blockstorageSnapshotCmd, &blockStorageSnapshotListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
