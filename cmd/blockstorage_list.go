package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageListItemOutput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Size  string `json:"size"`
	State string `json:"state"`
}

type blockstorageListOutput []blockstorageListItemOutput

func (o *blockstorageListOutput) ToJSON()  { output.JSON(o) }
func (o *blockstorageListOutput) ToText()  { output.Text(o) }
func (o *blockstorageListOutput) ToTable() { output.Table(o) }

type blockstorageListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
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
	TODO := context.TODO()

	resp, err := client.ListZones(TODO)
	if err != nil {
		return err
	}
	zones := resp.Zones

	if c.Zone != "" {
		endpoint, err := client.GetZoneAPIEndpoint(TODO, v3.ZoneName(c.Zone))
		if err != nil {
			return err
		}

		zones = []v3.Zone{{APIEndpoint: endpoint}}
	}

	output := make(blockstorageListOutput, 0)
	for _, zone := range zones {
		c := client.WithEndpoint(zone.APIEndpoint)

		resp, err := c.ListBlockStorageVolumes(TODO)
		if err != nil {
			// TODO: remove it once Block Storage is deployed in every zone.
			if strings.Contains(err.Error(), "Availability of the block storage volumes") {
				continue
			}

			return err
		}

		for _, volume := range resp.BlockStorageVolumes {
			output = append(output, blockstorageListItemOutput{
				ID:    volume.ID.String(),
				Name:  volume.Name,
				Zone:  string(zone.Name),
				Size:  humanize.IBytes(uint64(volume.Size)),
				State: string(volume.State),
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
