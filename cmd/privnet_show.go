package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type privnetShowOutput struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Zone      string   `json:"zone"`
	DHCP      string   `json:"dhcp"`
	Instances []string `json:"instances,omitempty"`
}

func (o *privnetShowOutput) Type() string { return "Private Network" }
func (o *privnetShowOutput) toJSON()      { outputJSON(o) }
func (o *privnetShowOutput) toText()      { outputText(o) }
func (o *privnetShowOutput) toTable()     { outputTable(o) }

func init() {
	privnetCmd.AddCommand(&cobra.Command{
		Use:   "show <privnet name | id>",
		Short: "Show a private network details",
		Long: fmt.Sprintf(`This command shows a Private Network details.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&privnetShowOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			return output(showPrivnet(args[0]))
		},
	})
}

func showPrivnet(name string) (outputter, error) {
	privnet, err := getNetwork(name, nil)
	if err != nil {
		return nil, err
	}

	out := privnetShowOutput{
		ID:   privnet.ID.String(),
		Name: privnet.Name,
		Zone: privnet.ZoneName,
		DHCP: dhcpRange(*privnet),
	}

	vms, err := privnetDetails(privnet)
	if err != nil {
		return nil, err
	}
	out.Instances = make([]string, len(vms))
	for i := range vms {
		out.Instances[i] = vms[i].Name
	}

	return &out, nil
}

func privnetDetails(network *egoscale.Network) ([]egoscale.VirtualMachine, error) {
	vms, err := cs.ListWithContext(gContext, &egoscale.VirtualMachine{
		ZoneID: network.ZoneID,
	})
	if err != nil {
		return nil, err
	}

	var vmsRes []egoscale.VirtualMachine
	for _, v := range vms {
		vm := v.(*egoscale.VirtualMachine)

		nic := vm.NicByNetworkID(*network.ID)
		if nic != nil {
			vmsRes = append(vmsRes, *vm)
		}
	}

	return vmsRes, nil
}
