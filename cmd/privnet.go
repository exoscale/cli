package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// privnetCmd represents the pn command
var privnetCmd = &cobra.Command{
	Use:   "privnet",
	Short: "Private networks management",
}

func getNetworkIDByName(cs *egoscale.Client, name, zone string) (string, error) {
	nets, err := cs.List(&egoscale.Network{Type: "Isolated", CanUseForDeploy: true, ZoneID: zone})
	if err != nil {
		log.Fatal(err)
	}

	res := ""
	match := 0
	for _, net := range nets {
		n := net.(*egoscale.Network)
		if strings.Compare(name, n.Name) == 0 || strings.Compare(name, n.ID) == 0 {
			res = n.ID
			match++
		}
	}
	switch match {
	case 0:
		return "", fmt.Errorf("Unable to find this private network")
	case 1:
		return res, nil
	default:
		return "", fmt.Errorf("Multiple private network found")

	}
}

func init() {
	RootCmd.AddCommand(privnetCmd)
}
