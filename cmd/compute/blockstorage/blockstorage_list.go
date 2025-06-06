package blockstorage

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageListItemOutput struct {
	ID    v3.UUID                    `json:"id"`
	Name  string                     `json:"name"`
	Zone  v3.ZoneName                `json:"zone"`
	Size  string                     `json:"size"`
	State v3.BlockStorageVolumeState `json:"state"`
}

type blockStorageListOutput []blockStorageListItemOutput

func (o *blockStorageListOutput) ToJSON()  { output.JSON(o) }
func (o *blockStorageListOutput) ToText()  { output.Text(o) }
func (o *blockStorageListOutput) ToTable() { output.Table(o) }

type blockStorageListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *blockStorageListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *blockStorageListCmd) CmdShort() string { return "List Block Storage Volumes" }

func (c *blockStorageListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Block Storage Volumes.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageListOutput{}), ", "))
}

func (c *blockStorageListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	ctx := exocmd.GContext

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

	output := make(blockStorageListOutput, 0)
	for _, zone := range zones {
		c := client.WithEndpoint(zone.APIEndpoint)

		resp, err := c.ListBlockStorageVolumes(ctx)
		if err != nil {
			// TODO(pej): remove it once Block Storage is deployed in every zone.
			if strings.Contains(err.Error(), "Availability of the block storage volumes") {
				continue
			}

			_, _ = fmt.Fprintf(os.Stderr,
				"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
			continue
		}

		for _, volume := range resp.BlockStorageVolumes {
			output = append(output, blockStorageListItemOutput{
				ID:    volume.ID,
				Name:  volume.Name,
				Zone:  zone.Name,
				Size:  fmt.Sprintf("%d GiB", volume.Size),
				State: volume.State,
			})
		}
	}

	return c.OutputFunc(&output, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(blockstorageCmd, &blockStorageListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
