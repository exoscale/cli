package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var eipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List elastic IP",
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"zone", "IP", "ID"})
		listIPs(zone, table)
		table.Render()
		return nil
	},
}

func listIPs(zone string, table *table.Table) {
	zReq := egoscale.IPAddress{}

	if zone != "" {
		var err error
		zReq.ZoneID, err = getZoneIDByName(cs, zone)
		if err != nil {
			log.Fatal(err)
		}
		zReq.IsElastic = true
		ips, err := cs.List(&zReq)
		if err != nil {
			log.Fatal(err)
		}

		for _, ipaddr := range ips {
			ip := ipaddr.(*egoscale.IPAddress)
			table.Append([]string{ip.ZoneName, ip.IPAddress.String(), ip.ID})
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
		listIPs(zID, table)
	}
}

func init() {
	eipListCmd.Flags().StringP("zone", "z", "", "Show IPs from given zone")
	eipCmd.AddCommand(eipListCmd)
}
