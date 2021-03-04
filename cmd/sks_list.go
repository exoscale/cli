package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
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
		for cluster := range res {
			out = append(out, cluster)
		}
	}()
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))
	err := forEachZone(sksClusterZones, func(zone string) error {
		list, err := cs.ListSKSClusters(ctx, zone)
		if err != nil {
			return fmt.Errorf("unable to list SKS clusters in zone %s: %v", zone, err)
		}

		for _, cluster := range list {
			res <- sksClusterListItemOutput{
				ID:   cluster.ID,
				Name: cluster.Name,
				Zone: zone,
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
	sksListCmd.Flags().StringP("zone", "z", "", "Zone to filter results to")
	sksCmd.AddCommand(sksListCmd)
}
