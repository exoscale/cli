package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type deployTargetShowOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Zone        string `json:"zone"`
}

func (o *deployTargetShowOutput) ToJSON()  { output.JSON(o) }
func (o *deployTargetShowOutput) ToText()  { output.Text(o) }
func (o *deployTargetShowOutput) ToTable() { output.Table(o) }

type deployTargetShowCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	DeployTarget string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"Deploy Target zone"`
}

func (c *deployTargetShowCmd) CmdAliases() []string { return GShowAlias }

func (c *deployTargetShowCmd) CmdShort() string { return "Show a Deploy Target details" }

func (c *deployTargetShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Deploy Target details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&deployTargetShowOutput{}), ", "))
}

func (c *deployTargetShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *deployTargetShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	dt, err := globalstate.EgoscaleClient.FindDeployTarget(ctx, c.Zone, c.DeployTarget)
	if err != nil {
		return fmt.Errorf("error retrieving Deploy Target: %w", err)
	}

	return c.OutputFunc(&deployTargetShowOutput{
		ID:          *dt.ID,
		Name:        *dt.Name,
		Description: utils.DefaultString(dt.Description, ""),
		Type:        *dt.Type,
		Zone:        c.Zone,
	}, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(deployTargetCmd, &deployTargetShowCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
