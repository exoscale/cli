package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type instancePoolItem struct {
	ID    string                     `json:"id"`
	Name  string                     `json:"name"`
	Zone  string                     `json:"zone"`
	Size  int                        `json:"size"`
	State egoscale.InstancePoolState `json:"state"`
}

type instancePoolListItemOutput []instancePoolItem

func (o *instancePoolListItemOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolListItemOutput) toText()  { outputText(o) }
func (o *instancePoolListItemOutput) toTable() { outputTable(o) }

type instancePoolFetchResult struct {
	instancePoolListItemOutput
	error
}

var instancePoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List instance pools",
	Long: fmt.Sprintf(`This command lists instance pools.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolItem{}), ", ")),
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zoneFlag, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zoneFlag = strings.ToLower(zoneFlag)

		var zones []egoscale.Zone
		if zoneFlag == "" {
			resp, err := cs.Request(egoscale.ListZones{})
			if err != nil {
				return err
			}
			zones = resp.(*egoscale.ListZonesResponse).Zone
		} else {
			zone, err := getZoneByName(zoneFlag)
			if err != nil {
				return err
			}
			zones = append(zones, *zone)
		}

		results := make(chan instancePoolFetchResult, len(zones))
		defer close(results)

		for _, zone := range zones {
			go getInstancePool(results, zone)
		}

		o := make(instancePoolListItemOutput, 0, len(zones))
		for range zones {
			result := <-results
			if result.error != nil {
				return err
			}

			o = append(o, result.instancePoolListItemOutput...)
		}

		return output(&o, nil)
	},
}

func getInstancePool(result chan instancePoolFetchResult, zone egoscale.Zone) {
	resp, err := cs.RequestWithContext(gContext, egoscale.ListInstancePools{
		ZoneID: zone.ID,
	})
	if err != nil {
		result <- instancePoolFetchResult{nil, err}
		return
	}
	r := resp.(*egoscale.ListInstancePoolsResponse)
	output := make(instancePoolListItemOutput, 0, r.Count)
	for _, i := range r.InstancePools {
		output = append(output, instancePoolItem{
			ID:    i.ID.String(),
			Name:  i.Name,
			Zone:  zone.Name,
			Size:  i.Size,
			State: i.State,
		})
	}

	result <- instancePoolFetchResult{output, nil}
}

func init() {
	instancePoolListCmd.Flags().StringP("zone", "z", "", "List Instance pool by zone")
	instancePoolCmd.AddCommand(instancePoolListCmd)
}
