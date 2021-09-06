package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type antiAffinityGroupShowOutput struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Instances   []string `json:"instances"`
}

func (o *antiAffinityGroupShowOutput) toJSON()  { outputJSON(o) }
func (o *antiAffinityGroupShowOutput) toText()  { outputText(o) }
func (o *antiAffinityGroupShowOutput) toTable() { outputTable(o) }

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
		strings.Join(outputterTemplateAnnotations(&antiAffinityGroupShowOutput{}), ", "))
}

func (c *antiAffinityGroupShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *antiAffinityGroupShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return output(showAntiAffinityGroup(gCurrentAccount.DefaultZone, c.AntiAffinityGroup))
}

func showAntiAffinityGroup(zone, x string) (outputter, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	antiAffinityGroup, err := cs.FindAntiAffinityGroup(ctx, zone, x)
	if err != nil {
		return nil, err
	}

	out := antiAffinityGroupShowOutput{
		ID:          *antiAffinityGroup.ID,
		Name:        *antiAffinityGroup.Name,
		Description: defaultString(antiAffinityGroup.Description, ""),
	}

	if antiAffinityGroup.InstanceIDs != nil {
		out.Instances = make([]string, len(*antiAffinityGroup.InstanceIDs))
		for i, id := range *antiAffinityGroup.InstanceIDs {
			instance, err := cs.GetInstance(ctx, zone, id)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve Compute instance %s: %v", id, err)
			}
			out.Instances[i] = *instance.Name
		}
	}

	return &out, nil
}

func init() {
	cobra.CheckErr(registerCLICommand(antiAffinityGroupCmd, &antiAffinityGroupShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
