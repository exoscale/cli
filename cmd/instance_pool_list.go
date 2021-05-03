package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolListItemOutput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Size  int64  `json:"size"`
	State string `json:"state"`
}

type instancePoolListOutput []instancePoolListItemOutput

func (o *instancePoolListOutput) toJSON()  { outputJSON(o) }
func (o *instancePoolListOutput) toText()  { outputText(o) }
func (o *instancePoolListOutput) toTable() { outputTable(o) }

var instancePoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Instance Pools",
	Long: fmt.Sprintf(`This command lists Instance Pools.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instancePoolListItemOutput{}), ", ")),
	Aliases: gListAlias,

	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone = strings.ToLower(zone)

		return output(listInstancePools(zone), nil)
	},
}

func listInstancePools(zone string) outputter {
	var zones []string

	if zone != "" {
		zones = []string{zone}
	} else {
		zones = allZones
	}

	out := make(instancePoolListOutput, 0)
	res := make(chan instancePoolListItemOutput)
	defer close(res)

	go func() {
		for instancePool := range res {
			out = append(out, instancePool)
		}
	}()
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	err := forEachZone(zones, func(zone string) error {
		list, err := cs.ListInstancePools(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list Instance Pools in zone %s: %v", zone, err)
		}

		for _, i := range list {
			res <- instancePoolListItemOutput{
				ID:    i.ID,
				Name:  i.Name,
				Zone:  zone,
				Size:  i.Size,
				State: i.State,
			}
		}

		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return &out
}

func init() {
	instancePoolListCmd.Flags().StringP("zone", "z", "", "Zone to filter results to")
	instancePoolCmd.AddCommand(instancePoolListCmd)
}
