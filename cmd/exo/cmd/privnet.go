package cmd

import (
	"fmt"
	"log"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// privnetCmd represents the pn command
var privnetCmd = &cobra.Command{
	Use:   "privnet",
	Short: "Private networks management",
}

func getNetworkIDByName(cs *egoscale.Client, name string) (*egoscale.Network, error) {
	nets, err := cs.List(&egoscale.Network{Type: "Isolated", CanUseForDeploy: true})
	if err != nil {
		log.Fatal(err)
	}

	var res *egoscale.Network
	match := 0
	for _, net := range nets {
		n := net.(*egoscale.Network)
		if name == n.Name || name == n.ID {
			res = n
			match++
		}
	}
	switch match {
	case 0:
		return nil, fmt.Errorf("Unable to find this private network")
	case 1:
		return res, nil
	default:
		return nil, fmt.Errorf("Multiple private network found")

	}
}

func init() {
	RootCmd.AddCommand(privnetCmd)
}
