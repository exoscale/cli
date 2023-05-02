package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

var (
	// allZones represents the list of known Exoscale zones, in case we need it without performing API lookup.
	allZones = []string{
		string(oapi.ZoneNameAtVie1),
		string(oapi.ZoneNameAtVie2),
		string(oapi.ZoneNameBgSof1),
		string(oapi.ZoneNameChDk2),
		string(oapi.ZoneNameChGva2),
		string(oapi.ZoneNameDeFra1),
		string(oapi.ZoneNameDeMuc1),
	}

	zoneHelp = "zone NAME|ID " + func() string {
		zonesList := "("

		for _, zone := range allZones {
			zonesList += zone + "|"
		}

		return zonesList[:len(zonesList)-1] + ")"
	}()
)

type zoneListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	zones, err := globalstate.EgoscaleClient.ListWithContext(gContext, &egoscale.Zone{})
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

func getZoneByNameOrID(name string) (*egoscale.Zone, error) {
	zone := &egoscale.Zone{}

	id, err := egoscale.ParseUUID(name)
	if err != nil {
		zone.Name = name
	} else {
		zone.ID = id
	}

	resp, err := globalstate.EgoscaleClient.GetWithContext(gContext, zone)
	if err != nil {
		if err == egoscale.ErrNotFound {
			return nil, fmt.Errorf("invalid zone %q", name)
		}
		return nil, err
	}

	return resp.(*egoscale.Zone), nil
}
