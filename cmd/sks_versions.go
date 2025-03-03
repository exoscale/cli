package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksClusterVersionsItemOutput struct {
	Version string `json:"version"`
}

type sksClusterVersionsOutput []sksClusterVersionsItemOutput

func (o *sksClusterVersionsOutput) ToJSON()  { output.JSON(o) }
func (o *sksClusterVersionsOutput) ToText()  { output.Text(o) }
func (o *sksClusterVersionsOutput) ToTable() { output.Table(o) }

type sksVersionsCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"versions"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone to filter results to"`
}

func (c *sksVersionsCmd) cmdAliases() []string { return gListAlias }

func (c *sksVersionsCmd) cmdShort() string { return "List supported SKS cluster versions" }

func (c *sksVersionsCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists supported SKS cluster versions.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksClusterVersionsItemOutput{}), ", "))
}

func (c *sksVersionsCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksVersionsCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	out := make(sksClusterVersionsOutput, 0)

	versions, err := client.ListSKSClusterVersions(ctx)
	if err != nil {
		return err
	}

	for _, v := range versions.SKSClusterVersions {
		_ = append(out, sksClusterVersionsItemOutput{Version: v})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(sksCmd, &sksVersionsCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
