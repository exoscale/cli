package cmd

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var eipCmd = &cobra.Command{
	Use:   "eip",
	Short: "Elastic IP management",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		fmt.Fprintln(os.Stderr,
			`**********************************************************************
The "exo eip" commands are deprecated and will be removed in a future
version, please use "exo compute elastic-ip" replacement commands.
**********************************************************************`)
		time.Sleep(3 * time.Second)
	},
	Hidden: true,
}

func getElasticIPByAddressOrID(v string) (*egoscale.IPAddress, error) {
	ip := net.ParseIP(v)
	id, _ := egoscale.ParseUUID(v)

	eips, err := globalstate.EgoscaleClient.ListWithContext(gContext, &egoscale.IPAddress{IsElastic: true})
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
