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
	Use:   "create <zone name | zone id>",
	Short: "Create EIP",
}

func runEIPCreateCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		eipCreateCmd.Usage()
		return
	}
	associateIPAddress(args[0])
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

	ipResp := resp.(*egoscale.AssociateIPAddressResponse)

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Zone", "IP", "ID"})

	table.Append([]string{ipResp.IPAddress.ZoneName, ipResp.IPAddress.IPAddress.String(), ipResp.IPAddress.ID})

	table.Render()

}

func init() {
	eipCreateCmd.Run = runEIPCreateCmd
	eipCmd.AddCommand(eipCreateCmd)
}
