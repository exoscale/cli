package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type deployTargetShowOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Zone        string `json:"zone"`
}

func (o *deployTargetShowOutput) toJSON()  { outputJSON(o) }
func (o *deployTargetShowOutput) toText()  { outputText(o) }
func (o *deployTargetShowOutput) toTable() { outputTable(o) }

type deployTargetShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	DeployTarget string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"Deploy Target zone"`
}

func (c *deployTargetShowCmd) cmdAliases() []string { return gShowAlias }

func (c *deployTargetShowCmd) cmdShort() string { return "Show a Deploy Target details" }

func (c *deployTargetShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Deploy Target details.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&deployTargetShowOutput{}), ", "))
}

func (c *deployTargetShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *deployTargetShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	dt, err := cs.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
	if err != nil {
		return fmt.Errorf("error retrieving Deploy Target: %w", err)
	}

	return c.outputFunc(&deployTargetShowOutput{
		ID:          *dt.ID,
		Name:        *dt.Name,
		Description: utils.DefaultString(dt.Description, ""),
		Type:        *dt.Type,
		Zone:        c.Zone,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(deployTargetCmd, &deployTargetShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
