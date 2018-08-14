package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// privnetCmd represents the pn command
var privnetCmd = &cobra.Command{
	Use:   "privnet",
	Short: "Private networks management",
}

func getNetworkByName(name string) (*egoscale.Network, error) {
	net := &egoscale.Network{
		Type:            "Isolated",
		CanUseForDeploy: true,
	}

	id, err := egoscale.ParseUUID(name)
	if err != nil {
		net.Name = name
	} else {
		net.ID = id
	}

	if err := cs.GetWithContext(gContext, net); err != nil {
		return nil, err
	}

	return net, err
}

func init() {
	RootCmd.AddCommand(privnetCmd)
}
