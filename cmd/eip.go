package cmd

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var eipCmd = &cobra.Command{
	Use:   "eip",
	Short: "Elastic IP management",
}

func getElasticIPByAddressOrID(v string) (*egoscale.IPAddress, error) {
	ip := net.ParseIP(v)
	id, _ := egoscale.ParseUUID(v)

	eips, err := cs.ListWithContext(gContext, &egoscale.IPAddress{IsElastic: true})
	if err != nil {
		return nil, err
	}

	for _, e := range eips {
		eip := e.(*egoscale.IPAddress)
		if (ip != nil && eip.IPAddress.Equal(ip)) || (id != nil && id.Equal(*eip.ID)) {
			return eip, nil
		}
	}

	return nil, fmt.Errorf("Elastic IP %q not found", v) // nolint
}

func init() {
	RootCmd.AddCommand(eipCmd)
}
