package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type privnetListItemOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Zone         string `json:"zone"`
	DHCP         string `json:"dhcp"`
	NumInstances int    `json:"num_instances" outputLabel:"Instances"`
}

type privnetListOutput []privnetListItemOutput

func (o *privnetListOutput) toJSON()  { outputJSON(o) }
func (o *privnetListOutput) toText()  { outputText(o) }
func (o *privnetListOutput) toTable() { outputTable(o) }

func init() {
	var privnetListCmd = &cobra.Command{
		Use:   "list",
		Short: "List private networks",
		Long: fmt.Sprintf(`This command lists existing Private Networks.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&privnetListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			zone, err := cmd.Flags().GetString("zone")
			if err != nil {
				return err
			}

			return output(listPrivnets(zone))
		},
	}

	privnetListCmd.Flags().StringP("zone", "z", "", "Show Private Networks only in specified zone")
	privnetCmd.AddCommand(privnetListCmd)
}

func listPrivnets(zone string) (outputter, error) {
	out := privnetListOutput{}

	zones, err := cs.ListWithContext(gContext, &egoscale.Zone{})
	if err != nil {
		return nil, err
	}

	for _, z := range zones {
		if zone != "" && z.(*egoscale.Zone).Name != zone {
			continue
		}

		req := egoscale.Network{
			ZoneID:          z.(*egoscale.Zone).ID,
			Type:            "Isolated",
			CanUseForDeploy: true,
		}

		privnets, err := cs.ListWithContext(gContext, &req)
		if err != nil {
			return nil, err
		}

		for _, p := range privnets {
			privnet := p.(*egoscale.Network)

			vms, err := privnetDetails(privnet)
			if err != nil {
				return nil, err
			}
			instances := make([]string, len(vms))
			for i := range vms {
				instances[i] = vms[i].Name
			}

			o := privnetListItemOutput{
				ID:           privnet.ID.String(),
				Name:         privnet.Name,
				Zone:         z.(*egoscale.Zone).Name,
				DHCP:         dhcpRange(*privnet),
				NumInstances: len(instances),
			}

			out = append(out, o)
		}
	}

	return &out, nil
}
