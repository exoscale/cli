package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var eipCreateCmd = &cobra.Command{
	Use:     "create [zone name | zone id]",
	Short:   "Create EIP",
	Aliases: gCreateAlias,
}

func runEIPCreateCmd(cmd *cobra.Command, args []string) {
	zone := gCurrentAccount.DefaultZone
	if len(args) >= 1 {
		zone = args[0]
	}
	associateIPAddress(zone)
}

func associateIPAddress(name string) {
	ipReq := egoscale.AssociateIPAddress{}

	var err error
	ipReq.ZoneID, err = getZoneIDByName(cs, name)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := cs.Request(&ipReq)
	if err != nil {
		log.Fatal(err)
	}

	ipResp := resp.(*egoscale.IPAddress)

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Zone", "IP", "ID"})

	table.Append([]string{ipResp.ZoneName, ipResp.IPAddress.String(), ipResp.ID})

	table.Render()

}

func init() {
	eipCreateCmd.Run = runEIPCreateCmd
	eipCmd.AddCommand(eipCreateCmd)
}
