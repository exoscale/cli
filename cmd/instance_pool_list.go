package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolListItemOutput struct {
	ID    string                     `json:"id"`
	Name  string                     `json:"name"`
	Zone  string                     `json:"zone"`
	Size  int                        `json:"size"`
	State egoscale.InstancePoolState `json:"state"`
}

type instancePoolListOutput []instancePoolListItemOutput

func (o *instancePoolListOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolListOutput) toText()  { outputText(o) }
func (o *instancePoolListOutput) toTable() { outputTable(o) }

var instancePoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List instance pools",
	Long: fmt.Sprintf(`This command lists instance pools.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolListItemOutput{}), ", ")),
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone = strings.ToLower(zone)

		return output(listInstancePools(zone))
	},
}

func listInstancePools(zone string) (outputter, error) {
	var (
		zonesIndex        = make(map[string]egoscale.Zone)
		instancePoolZones = make([]string, 0)
	)

	// We have to index existing zones per name in advance, as forEachZone()
	// expects a zone name but we'll need the zone UUID to perform the CS-style
	// calls in the callback function.
	resp, err := cs.RequestWithContext(gContext, egoscale.ListZones{})
	if err != nil {
		return nil, err
	}
	for _, z := range resp.(*egoscale.ListZonesResponse).Zone {
		if zone != "" && z.Name != zone {
			continue
		}

		zonesIndex[z.Name] = z
		instancePoolZones = append(instancePoolZones, z.Name)
	}

	out := make(instancePoolListOutput, 0)
	res := make(chan instancePoolListItemOutput)
	defer close(res)

	go func() {
		for instancePool := range res {
			out = append(out, instancePool)
		}
	}()
	err = forEachZone(instancePoolZones, func(zone string) error {
		resp, err := cs.RequestWithContext(gContext, egoscale.ListInstancePools{ZoneID: zonesIndex[zone].ID})
		if err != nil {
			return err
		}

		for _, i := range resp.(*egoscale.ListInstancePoolsResponse).InstancePools {
			res <- instancePoolListItemOutput{
				ID:    i.ID.String(),
				Name:  i.Name,
				Zone:  zone,
				Size:  i.Size,
				State: i.State,
			}
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return &out, nil
}

func init() {
	instancePoolListCmd.Flags().StringP("zone", "z", "", "Zone to filter results to")
	instancePoolCmd.AddCommand(instancePoolListCmd)
}
