package cmd

import (
	"net"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var eipDeleteCmd = &cobra.Command{
	Use:     "delete <ip | eip id>",
	Short:   "Delete EIP",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		return deleteEip(args[0])
	},
}

func deleteEip(ip string) error {
	addrReq := &egoscale.DisassociateIPAddress{}

	ipAddr := net.ParseIP(ip)

	if ipAddr == nil {
		addrReq.ID = ip
	} else {
		req := &egoscale.IPAddress{IPAddress: ipAddr, IsElastic: true}
		if err := cs.Get(req); err != nil {
			return err
		}
		addrReq.ID = req.ID
	}

	_, err := cs.Request(addrReq)
	return err
}

func init() {
	eipCmd.AddCommand(eipDeleteCmd)
}
