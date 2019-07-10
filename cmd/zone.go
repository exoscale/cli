package cmd

import (
	"os"

	"github.com/exoscale/cli/table"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

const (
	zoneHelp = "<zone name | id> (ch-dk-2|ch-gva-2|at-vie-1|de-fra-1|bg-sof-1|de-muc-1)"
)

// zoneCmd represents the zone command
var zoneCmd = &cobra.Command{
	Use:   "zone",
	Short: "List all available zones",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listZones()
	},
}

func listZones() error {
	zones, err := cs.ListWithContext(gContext, &egoscale.Zone{})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "ID"})

	for _, zone := range zones {
		z := zone.(*egoscale.Zone)
		table.Append([]string{z.Name, z.ID.String()})
	}
	table.Render()
	return nil
}

func getZoneIDByName(name string) (*egoscale.UUID, error) {
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

	return resp.(*egoscale.Zone).ID, nil
}

func init() {
	RootCmd.AddCommand(zoneCmd)
}
