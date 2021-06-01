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

type sksNodepoolListCmd struct {
	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *sksNodepoolListCmd) cmdAliases() []string { return gListAlias }

func (c *sksNodepoolListCmd) cmdShort() string { return "List SKS cluster Nodepools" }

func (c *sksNodepoolListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists SKS cluster Nodepools.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksNodepoolListItemOutput{}), ", "))
}

func (c *sksNodepoolListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksNodepoolListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zones[0]))

	out := make(sksNodepoolListOutput, 0)
	res := make(chan sksNodepoolListItemOutput)
	defer close(res)

	go func() {
		for cluster := range res {
			out = append(out, cluster)
		}
	}()
	err := forEachZone(zones, func(zone string) error {
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

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(sksNodepoolCmd, &sksNodepoolListCmd{}))
}
