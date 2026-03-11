package deployment

import (
	"fmt"
	"sort"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type InstanceTypeListItemOutput struct {
	Family     string `json:"family"`
	Authorized bool   `json:"authorized"`
	Zone       string `json:"zone"`
}

type InstanceTypeListOutput []InstanceTypeListItemOutput

func (o *InstanceTypeListOutput) ToJSON()  { output.JSON(o) }
func (o *InstanceTypeListOutput) ToText()  { output.Text(o) }
func (o *InstanceTypeListOutput) ToTable() { output.Table(o) }

type InstanceTypeListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"instance-type"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *InstanceTypeListCmd) CmdAliases() []string { return nil }
func (c *InstanceTypeListCmd) CmdShort() string     { return "List AI instance types" }
func (c *InstanceTypeListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI instance types.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceTypeListOutput{}), ", "))
}
func (c *InstanceTypeListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *InstanceTypeListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	zones, err := utils.AllZonesV3(ctx, client, c.Zone)
	if err != nil {
		return err
	}

	out := make(InstanceTypeListOutput, 0)
	err = utils.ForEveryZone(zones, func(zone v3.Zone) error {
		zoneClient := client.WithEndpoint(zone.APIEndpoint)
		resp, err := zoneClient.ListAIInstanceTypes(ctx)
		if err != nil {
			return err
		}

		for _, it := range resp.InstanceTypes {
			authorized := false
			if it.Authorized != nil {
				authorized = *it.Authorized
			}
			out = append(out, InstanceTypeListItemOutput{
				Family:     it.Family,
				Authorized: authorized,
				Zone:       string(zone.Name),
			})
		}

		return nil
	})

	sortInstanceTypeListOutput(out)

	return c.OutputFunc(&out, err)
}

// sortInstanceTypeListOutput sorts by zone then by family alphabetically.
func sortInstanceTypeListOutput(out InstanceTypeListOutput) {
	sort.Slice(out, func(i, j int) bool {
		if out[i].Zone < out[j].Zone {
			return true
		}
		if out[i].Zone > out[j].Zone {
			return false
		}
		return out[i].Family < out[j].Family
	})
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &InstanceTypeListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
