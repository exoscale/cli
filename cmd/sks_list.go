package cmd

import (
	"fmt"
	"os"
	"strings"

	apiv2 "github.com/exoscale/egoscale/api/v2"
	"github.com/spf13/cobra"
)

type sksClusterListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Zone string `json:"zone"`
}

type sksClusterListOutput []sksClusterListItemOutput

func (o *sksClusterListOutput) toJSON()  { outputJSON(o) }
func (o *sksClusterListOutput) toText()  { outputText(o) }
func (o *sksClusterListOutput) toTable() { outputTable(o) }

var sksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SKS clusters",
	Long: fmt.Sprintf(`This command lists SKS clusters.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksClusterListItemOutput{}), ", ")),
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone = strings.ToLower(zone)

		return output(listSKSClusters(zone), nil)
	},
}

func listSKSClusters(zone string) outputter {
	var sksClusterZones []string

	if zone != "" {
		sksClusterZones = []string{zone}
	} else {
		sksClusterZones = zones
	}

	out := make(sksClusterListOutput, 0)
	res := make(chan sksClusterListItemOutput)
	defer close(res)

	go func() {
		for nlb := range res {
			out = append(out, nlb)
		}
	}()
	ctx := apiv2.WithEndpoint(gContext, apiv2.NewReqEndpoint(gCurrentAccount.Environment, zone))
	err := forEachZone(sksClusterZones, func(zone string) error {
		list, err := cs.ListSKSClusters(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list SKS clusters in zone %s: %v", zone, err)
		}

		for _, nlb := range list {
			res <- sksClusterListItemOutput{
				ID:   nlb.ID,
				Name: nlb.Name,
				Zone: zone,
			}
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"warning: errors during listing, results might be incomplete.\n%s\n", err) // nolint:golint
	}

	return &out
}

func init() {
	sksListCmd.Flags().StringP("zone", "z", "", "Zone to filter results to")
	sksCmd.AddCommand(sksListCmd)
}
