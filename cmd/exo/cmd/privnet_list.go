package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var privnetListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List private networks",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"zone", "Name", "ID"})
		listPrivnets(zone, table)
		table.Render()
		return nil
	},
}

func listPrivnets(zone string, table *table.Table) {
	pnReq := &egoscale.Network{}

	if zone != "" {
		var err error
		pnReq.Type = "Isolated"
		pnReq.ZoneID, err = getZoneIDByName(cs, zone)
		if err != nil {
			log.Fatal(err)
		}
		pnReq.CanUseForDeploy = true
		pns, err := cs.List(pnReq)
		if err != nil {
			log.Fatal(err)
		}

		for _, pNet := range pns {
			pn := pNet.(*egoscale.Network)
			table.Append([]string{pn.ZoneName, pn.Name, pn.ID})
		}
		return
	}

	zones := &egoscale.Zone{}
	zs, err := cs.List(zones)
	if err != nil {
		log.Fatal(err)
	}

	for _, z := range zs {
		zID := z.(*egoscale.Zone).Name
		listPrivnets(zID, table)
	}
}

func init() {
	privnetListCmd.Flags().StringP("zone", "z", "", "Show Private Network from given zone")
	privnetCmd.AddCommand(privnetListCmd)
}
