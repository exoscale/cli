package cmd

import (
	"fmt"
	"os"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type sksNodepoolListItemOutput struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Cluster string `json:"cluster"`
	Size    int64  `json:"size"`
	State   string `json:"state"`
	Zone    string `json:"zone"`
}

type sksNodepoolListOutput []sksNodepoolListItemOutput

func (o *sksNodepoolListOutput) toJSON()  { outputJSON(o) }
func (o *sksNodepoolListOutput) toText()  { outputText(o) }
func (o *sksNodepoolListOutput) toTable() { outputTable(o) }

var sksNodepoolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SKS cluster Nodepools",
	Long: fmt.Sprintf(`This command lists SKS cluster Nodepools.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksNodepoolListItemOutput{}), ", ")),
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}
		zone = strings.ToLower(zone)

		return output(listSKSNodepools(zone), nil)
	},
}

func listSKSNodepools(zone string) outputter {
	var sksClusterZones []string

	if zone != "" {
		sksClusterZones = []string{zone}
	} else {
		sksClusterZones = allZones
	}

	out := make(sksNodepoolListOutput, 0)
	res := make(chan sksNodepoolListItemOutput)
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
			for _, np := range cluster.Nodepools {
				res <- sksNodepoolListItemOutput{
					ID:      np.ID,
					Name:    np.Name,
					Cluster: cluster.Name,
					Size:    np.Size,
					State:   np.State,
					Zone:    zone,
				}
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
	sksNodepoolListCmd.Flags().StringP("zone", "z", "", "Zone to filter results to")
	sksNodepoolCmd.AddCommand(sksNodepoolListCmd)
}
