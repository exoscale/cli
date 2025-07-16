package sks

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

// TODO: full v3 migration is blocked by
// https://app.shortcut.com/exoscale/story/122943/bug-in-egoscale-v3-listsksclusterdeprecatedresources

type sksListDeprecatedResourcesOutput []v3.SKSClusterDeprecatedResource

func (o *sksListDeprecatedResourcesOutput) ToJSON()  { output.JSON(o) }
func (o *sksListDeprecatedResourcesOutput) ToText()  { output.Text(o) }
func (o *sksListDeprecatedResourcesOutput) ToTable() { output.Table(o) }

type sksDeprecatedResourcesCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"deprecated-resources"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Zone    string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksDeprecatedResourcesCmd) CmdAliases() []string { return []string{"dr"} }

func (c *sksDeprecatedResourcesCmd) CmdShort() string {
	return "List resources that will be deprecated in a futur release of Kubernetes for an SKS cluster"
}

func (c *sksDeprecatedResourcesCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists SKS cluster Nodepools.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksListDeprecatedResourcesOutput{}), ", "))
}

func (c *sksDeprecatedResourcesCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksDeprecatedResourcesCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	clusters, err := client.ListSKSClusters(ctx)
	if err != nil {
		return err
	}

	cluster, err := clusters.FindSKSCluster(c.Cluster)
	if err != nil {
		return err
	}

	deprecatedResources, err := client.ListSKSClusterDeprecatedResources(ctx, cluster.ID)
	if err != nil {
		return fmt.Errorf("error retrieving deprecated resources: %w", err)
	}

	// deprecatedResources.
	out := make(sksListDeprecatedResourcesOutput, 0)

	for _, t := range deprecatedResources {
		out = append(out, t)
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksDeprecatedResourcesCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
