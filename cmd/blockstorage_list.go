package cmd

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageListItemOutput struct {
	ID    v3.UUID                    `json:"id"`
	Name  string                     `json:"name"`
	Zone  v3.ZoneName                `json:"zone"`
	Size  string                     `json:"size"`
	State v3.BlockStorageVolumeState `json:"state"`
}

type blockstorageListOutput []blockstorageListItemOutput

func (o *blockstorageListOutput) ToJSON()  { output.JSON(o) }
func (o *blockstorageListOutput) ToText()  { output.Text(o) }
func (o *blockstorageListOutput) ToTable() { output.Table(o) }

type blockstorageListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *blockstorageListCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageListCmd) cmdShort() string { return "List Block Storage Volumes" }

func (c *blockstorageListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Block Storage Volumes.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageListOutput{}), ", "))
}

func (c *blockstorageListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageListCmd) cmdRun(_ *cobra.Command, _ []string) error {
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

	output := make(blockstorageListOutput, 0)
	for _, zone := range zones {
		c := client.WithEndpoint(zone.APIEndpoint)

		resp, err := c.ListBlockStorageVolumes(ctx)
		if err != nil {
			// TODO: remove it once Block Storage is deployed in every zone.
			if strings.Contains(err.Error(), "Availability of the block storage volumes") {
				continue
			}

			return err
		}

		for _, volume := range resp.BlockStorageVolumes {
			output = append(output, blockstorageListItemOutput{
				ID:    volume.ID,
				Name:  volume.Name,
				Zone:  zone.Name,
				Size:  humanize.IBytes(uint64(volume.Size)),
				State: volume.State,
			})
		}
	}

	return c.outputFunc(&output, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
