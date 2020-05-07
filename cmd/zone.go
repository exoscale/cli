package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

const (
	zoneHelp = "<zone name | id> (ch-dk-2|ch-gva-2|at-vie-1|de-fra-1|bg-sof-1|de-muc-1)"
)

// zones represents the list of known Exoscale zones, in case we need it without performing API lookup.
var zones = []string{
	"at-vie-1",
	"bg-sof-1",
	"ch-dk-2",
	"ch-gva-2",
	"de-fra-1",
	"de-muc-1",
}

type zoneListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type zoneListOutput []zoneListItemOutput

func (o *zoneListOutput) toJSON()  { outputJSON(o) }
func (o *zoneListOutput) toText()  { outputText(o) }
func (o *zoneListOutput) toTable() { outputTable(o) }

func (o zoneListOutput) Len() int           { return len(o) }
func (o zoneListOutput) Swap(x, y int)      { o[x], o[y] = o[y], o[x] }
func (o zoneListOutput) Less(x, y int) bool { return o[x].Name < o[y].Name }

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "zone",
		Short: "List all available zones",
		Long: fmt.Sprintf(`This command lists available Exoscale zones.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&zoneListOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listZones())
		},
	})
}

func listZones() (outputter, error) {
	zones, err := cs.ListWithContext(gContext, &egoscale.Zone{})
	if err != nil {
		return nil, err
	}

	out := zoneListOutput{}

	for _, key := range zones {
		zone := key.(*egoscale.Zone)

		out = append(out, zoneListItemOutput{
			ID:   zone.ID.String(),
			Name: zone.Name,
		})
	}

	sort.Sort(out)

	return &out, nil
}

func getZoneByName(name string) (*egoscale.Zone, error) {
	zone := &egoscale.Zone{}

	id, err := egoscale.ParseUUID(name)
	if err != nil {
		zone.Name = name
	} else {
		zone.ID = id
	}

	resp, err := cs.GetWithContext(gContext, zone)
	if err != nil {
		return nil, err
	}

	return resp.(*egoscale.Zone), nil
}
