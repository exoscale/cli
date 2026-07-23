package vpc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcSubnetListItemOutput struct {
	ID            v3.UUID `json:"id"`
	Name          string  `json:"name"`
	IPv4Block     string  `json:"ipv4_block" outputLabel:"IPv4 Block"`
	AddressFamily string  `json:"address_family"`
	Description   string  `json:"description"`
}

type vpcSubnetListOutput []vpcSubnetListItemOutput

func (o *vpcSubnetListOutput) ToJSON()  { output.JSON(o) }
func (o *vpcSubnetListOutput) ToText()  { output.Text(o) }
func (o *vpcSubnetListOutput) ToTable() { output.Table(o) }

type vpcSubnetListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	VPC string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcSubnetListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *vpcSubnetListCmd) CmdShort() string { return "List VPC Subnets" }

func (c *vpcSubnetListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists the Subnets of a Virtual Private Cloud.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&vpcSubnetListItemOutput{}), ", "))
}

func (c *vpcSubnetListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcSubnetListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	out, err := c.list()
	if err != nil {
		return err
	}

	return c.OutputFunc(out, nil)
}

// list resolves the VPC and returns its Subnets to display.
func (c *vpcSubnetListCmd) list() (*vpcSubnetListOutput, error) {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return nil, err
	}

	vpcEntry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return nil, err
	}

	resp, err := client.ListSubnets(ctx, vpcEntry.ID)
	if err != nil {
		return nil, err
	}

	out := make(vpcSubnetListOutput, 0, len(resp.Subnets))
	for _, s := range resp.Subnets {
		out = append(out, vpcSubnetListItemOutput{
			ID:            s.ID,
			Name:          s.Name,
			IPv4Block:     s.Ipv4Block,
			AddressFamily: string(s.Addressfamily),
			Description:   s.Description,
		})
	}

	return &out, nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcSubnetCmd, &vpcSubnetListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
