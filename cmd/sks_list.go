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

type sksListCmd struct {
	_ bool `cli-cmd:"list"`

	Zone string `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *sksListCmd) cmdAliases() []string { return gListAlias }

func (c *sksListCmd) cmdShort() string { return "List SKS clusters" }

func (c *sksListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists SKS clusters.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&sksClusterListItemOutput{}), ", "))
}

func (c *sksListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	var zones []string

	if c.Zone != "" {
		zones = []string{c.Zone}
	} else {
		zones = allZones
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	out := make(sksClusterListOutput, 0)
	res := make(chan sksClusterListItemOutput)
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
			res <- sksClusterListItemOutput{
				ID:   *cluster.ID,
				Name: *cluster.Name,
				Zone: zone,
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
	cobra.CheckErr(registerCLICommand(sksCmd, &sksListCmd{}))
}
