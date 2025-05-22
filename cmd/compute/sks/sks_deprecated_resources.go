package sks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

// TODO: full v3 migration is blocked by
// https://app.shortcut.com/exoscale/story/122943/bug-in-egoscale-v3-listsksclusterdeprecatedresources

type sksListDeprecatedResourcesItemOutput struct {
	Group          string `json:"group"`
	Version        string `json:"version"`
	Resource       string `json:"resource"`
	SubResource    string `json:"subresource"`
	RemovedRelease string `json:"removed_release"`
}

type sksListDeprecatedResourcesOutput []sksListDeprecatedResourcesItemOutput

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
		strings.Join(output.TemplateAnnotations(&sksListDeprecatedResourcesItemOutput{}), ", "))
}

func emptyIfNil(inp *string) string {
	if inp == nil {
		return ""
	}

	return *inp
}

func (c *sksDeprecatedResourcesCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksDeprecatedResourcesCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	cluster, err := globalstate.EgoscaleClient.FindSKSCluster(ctx, c.Zone, c.Cluster)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	deprecatedResources, err := globalstate.EgoscaleClient.ListSKSClusterDeprecatedResources(
		ctx,
		c.Zone,
		cluster,
	)
	if err != nil {
		return fmt.Errorf("error retrieving deprecated resources: %w", err)
	}

	out := make(sksListDeprecatedResourcesOutput, 0)

	for _, t := range deprecatedResources {
		out = append(out, sksListDeprecatedResourcesItemOutput{
			Group:          emptyIfNil(t.Group),
			RemovedRelease: emptyIfNil(t.RemovedRelease),
			Resource:       emptyIfNil(t.Resource),
			SubResource:    emptyIfNil(t.SubResource),
			Version:        emptyIfNil(t.Version),
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksDeprecatedResourcesCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
