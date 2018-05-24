package cmd

import (
	"log"
	"net"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var eipDeleteCmd = &cobra.Command{
	Use:   "delete <ip | eip id>",
	Short: "Delete EIP",
}

func eipDeleteRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		eipDeleteCmd.Usage()
		return
	}
	deleteEip(args[0])
}

func deleteEip(ip string) {
	addrReq := &egoscale.DisassociateIPAddress{}

	ipAddr := net.ParseIP(ip)

	if ipAddr == nil {
		addrReq.ID = ip
	} else {
		req := &egoscale.IPAddress{IPAddress: ipAddr, IsElastic: true}
		if err := cs.Get(req); err != nil {
			log.Fatal(err)
		}
		addrReq.ID = req.ID
	}

	_, err := cs.Request(addrReq)
	if err != nil {
		log.Fatal(err)
	}

}

func init() {
	eipDeleteCmd.Run = eipDeleteRun
	eipCmd.AddCommand(eipDeleteCmd)
}
