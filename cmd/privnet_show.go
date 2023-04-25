package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type privnetShowOutput struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Zone        string   `json:"zone"`
	DHCP        string   `json:"dhcp"`
	Instances   []string `json:"instances,omitempty"`
}

func (o *privnetShowOutput) Type() string { return "Private Network" }
func (o *privnetShowOutput) toJSON()      { output.JSON(o) }
func (o *privnetShowOutput) toText()      { output.Text(o) }
func (o *privnetShowOutput) toTable()     { output.Table(o) }

func init() {
	privnetCmd.AddCommand(&cobra.Command{
		Use:   "show NAME|ID",
		Short: "Show a Private Network details",
		Long: fmt.Sprintf(`This command shows a Private Network details.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&privnetShowOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			privnet, err := getNetwork(args[0], nil)
			if err != nil {
				return err
			}

			return printOutput(showPrivnet(privnet))
		},
	})
}

func showPrivnet(privnet *egoscale.Network) (output.Outputter, error) {
	out := privnetShowOutput{
		ID:          privnet.ID.String(),
		Name:        privnet.Name,
		Description: privnet.DisplayText,
		Zone:        privnet.ZoneName,
		DHCP:        dhcpRange(*privnet),
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
	vms, err := globalstate.GlobalEgoscaleClient.ListWithContext(gContext, &egoscale.VirtualMachine{
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
