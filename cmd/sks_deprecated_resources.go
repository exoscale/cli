package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"deprecated-resources"`

	Cluster string `cli-arg:"#" cli-usage:"CLUSTER-NAME|ID"`
	Zone    string `cli-short:"z" cli-usage:"SKS cluster zone"`
}

func (c *sksDeprecatedResourcesCmd) cmdAliases() []string { return []string{"dr"} }

func (c *sksDeprecatedResourcesCmd) cmdShort() string {
	return "List resources that will be deprecated in a futur release of Kubernetes for an SKS cluster"
}

func (c *sksDeprecatedResourcesCmd) cmdLong() string {
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

func (c *sksDeprecatedResourcesCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksDeprecatedResourcesCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

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

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksDeprecatedResourcesCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
