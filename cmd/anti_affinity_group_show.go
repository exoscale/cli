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

type antiAffinityGroupShowOutput struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Instances   []string `json:"instances"`
}

func (o *antiAffinityGroupShowOutput) ToJSON()  { output.JSON(o) }
func (o *antiAffinityGroupShowOutput) ToText()  { output.Text(o) }
func (o *antiAffinityGroupShowOutput) ToTable() { output.Table(o) }

type antiAffinityGroupShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	AntiAffinityGroup string `cli-arg:"#" cli-usage:"NAME|ID"`
}

func (c *antiAffinityGroupShowCmd) cmdAliases() []string { return gShowAlias }

func (c *antiAffinityGroupShowCmd) cmdShort() string {
	return "Show an Anti-Affinity Group details"
}

func (c *antiAffinityGroupShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance Anti-Affinity Group details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&antiAffinityGroupShowOutput{}), ", "))
}

func (c *antiAffinityGroupShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	antiAffinityGroup, err := globalstate.EgoscaleClient.FindAntiAffinityGroup(ctx, zone, c.AntiAffinityGroup)
	if err != nil {
		return err
	}

	out := antiAffinityGroupShowOutput{
		ID:          *antiAffinityGroup.ID,
		Name:        *antiAffinityGroup.Name,
		Description: utils.DefaultString(antiAffinityGroup.Description, ""),
	}

	if antiAffinityGroup.InstanceIDs != nil {
		out.Instances = make([]string, len(*antiAffinityGroup.InstanceIDs))
		for i, id := range *antiAffinityGroup.InstanceIDs {
			instance, err := globalstate.EgoscaleClient.GetInstance(ctx, zone, id)
			if err != nil {
				return fmt.Errorf("unable to retrieve Compute instance %s: %w", id, err)
			}
			out.Instances[i] = *instance.Name
		}
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
