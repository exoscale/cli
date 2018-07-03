package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"

	"github.com/spf13/cobra"
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
	zones, err := cs.List(&egoscale.Zone{})
	if err != nil {
		return err
	}

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "ID"})

	for _, zone := range zones {
		z := zone.(*egoscale.Zone)
		table.Append([]string{z.Name, z.ID})
	}
	table.Render()
	return nil
}

func getZoneIDByName(cs *egoscale.Client, name string) (string, error) {

	zoneReq := egoscale.Zone{}

	zones, err := cs.List(&zoneReq)
	if err != nil {
		return "", err
	}

	keywords := []string{}

	for _, zone := range zones {
		z := zone.(*egoscale.Zone)
		if name == z.ID {
			return z.ID, nil
		}
		if strings.Contains(strings.ToLower(z.Name), strings.ToLower(name)) {
			keywords = append(keywords, z.ID)
		}
	}

	if len(keywords) > 1 {
		return "", fmt.Errorf("More than one zone found")
	}
	if len(keywords) == 1 {
		return keywords[0], nil
	}

	return "", fmt.Errorf("Zone not found")
}

func init() {
	RootCmd.AddCommand(zoneCmd)
}
