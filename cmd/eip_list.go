package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type eipListItemOutput struct {
	Description string   `json:"description"`
	ID          string   `json:"id"`
	Zone        string   `json:"zone"`
	IPAddress   string   `json:"ip_address"`
	Managed     bool     `json:"managed"`
	Instances   []string `json:"instances,omitempty"`
}

type eipListOutput []eipListItemOutput

func (o *eipListOutput) toJSON()  { outputJSON(o) }
func (o *eipListOutput) toText()  { outputText(o) }
func (o *eipListOutput) toTable() { outputTable(o) }

func init() {
	var eipListCmd = &cobra.Command{
		Use:   "list",
		Short: "List elastic IP",
		Long: fmt.Sprintf(`This command lists existing Elastic IP addresses.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&eipListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			zone, err := cmd.Flags().GetString("zone")
			if err != nil {
				return err
			}

			return output(listEIP(zone))
		},
	}

	eipListCmd.Flags().StringP("zone", "z", "", "Show IPs from given zone")
	eipCmd.AddCommand(eipListCmd)
}

func listEIP(zone string) (outputter, error) {
	out := eipListOutput{}

	zones, err := cs.ListWithContext(gContext, &egoscale.Zone{})
	if err != nil {
		return nil, err
	}

	for _, z := range zones {
		if zone != "" && z.(*egoscale.Zone).Name != zone {
			continue
		}

		req := egoscale.IPAddress{
			ZoneID:    z.(*egoscale.Zone).ID,
			IsElastic: true,
		}

		ips, err := cs.ListWithContext(gContext, &req)
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			eip := ip.(*egoscale.IPAddress)

			_, vms, err := eipDetails(eip.ID)
			if err != nil {
				return nil, err
			}
			instances := make([]string, len(vms))
			for i := range vms {
				instances[i] = vms[i].Name
			}

			o := eipListItemOutput{
				Description: eip.Description,
				ID:          eip.ID.String(),
				IPAddress:   eip.IPAddress.String(),
				Zone:        z.(*egoscale.Zone).Name,
				Instances:   instances,
			}
			if eip.Healthcheck != nil {
				o.Managed = true
			}

			out = append(out, o)
		}
	}

	return &out, nil
}
