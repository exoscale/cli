package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type zoneListItemOutput struct {
	Name v3.ZoneName `json:"name"`
}

type zoneListOutput []zoneListItemOutput

func (o *zoneListOutput) ToJSON()  { output.JSON(o) }
func (o *zoneListOutput) ToText()  { output.Text(o) }
func (o *zoneListOutput) ToTable() { output.Table(o) }

func (o zoneListOutput) Len() int           { return len(o) }
func (o zoneListOutput) Swap(x, y int)      { o[x], o[y] = o[y], o[x] }
func (o zoneListOutput) Less(x, y int) bool { return o[x].Name < o[y].Name }

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:     "zone",
		Aliases: []string{"zones"},
		Short:   "List all available zones",
		Long: fmt.Sprintf(`This command lists available Exoscale zones.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&zoneListOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(listZones())
		},
	})
}

func listZones() (output.Outputter, error) {

	ctx := gContext
	client := globalstate.EgoscaleV3Client

	zones, err := client.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	out := zoneListOutput{}

	for _, zone := range zones.Zones {

		out = append(out, zoneListItemOutput{
			Name: zone.Name,
		})
	}

	sort.Sort(out)

	return &out, nil
}
