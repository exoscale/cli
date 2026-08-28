package vpc

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcSubnetItemOutput struct {
	ID        v3.UUID `json:"id"`
	Name      string  `json:"name"`
	IPv4Block string  `json:"ipv4_block"`
}

type vpcShowOutput struct {
	ID          v3.UUID               `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Zone        v3.ZoneName           `json:"zone"`
	CreatedAt   string                `json:"created_at"`
	Labels      map[string]string     `json:"labels"`
	Subnets     []vpcSubnetItemOutput `json:"subnets"`
}

func (o *vpcShowOutput) ToJSON() { output.JSON(o) }
func (o *vpcShowOutput) ToText() { output.Text(o) }
func (o *vpcShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"VPC"})
	defer t.Render()

	t.Append([]string{"ID", o.ID.String()})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", string(o.Zone)})
	t.Append([]string{"Created At", o.CreatedAt})
	t.Append([]string{"Labels", formatLabels(o.Labels)})
	t.Append([]string{"Subnets", formatSubnets(o.Subnets)})
}

type vpcShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	VPC string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *vpcShowCmd) CmdShort() string { return "Show a VPC details" }

func (c *vpcShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Virtual Private Cloud details.

Supported output template annotations for VPC: %s

Supported output template annotations for VPC subnets: %s`,
		strings.Join(output.TemplateAnnotations(&vpcShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&vpcSubnetItemOutput{}), ", "))
}

func (c *vpcShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	entry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	vpc, err := client.GetVpc(ctx, entry.ID)
	if err != nil {
		return err
	}

	out := vpcShowOutput{
		ID:          vpc.ID,
		Name:        vpc.Name,
		Description: vpc.Description,
		Zone:        c.Zone,
		CreatedAt:   vpc.CreatedAT.String(),
		Labels:      vpc.Labels,
		Subnets:     []vpcSubnetItemOutput{},
	}

	subnets, err := client.ListSubnets(ctx, vpc.ID)
	if err != nil {
		return fmt.Errorf("unable to list Subnets of VPC %s: %w", vpc.ID, err)
	}

	for _, s := range subnets.Subnets {
		out.Subnets = append(out.Subnets, vpcSubnetItemOutput{
			ID:        s.ID,
			Name:      s.Name,
			IPv4Block: s.Ipv4Block,
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &vpcShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}

func formatSubnets(subnets []vpcSubnetItemOutput) string {
	if len(subnets) == 0 {
		return "-"
	}

	buf := bytes.NewBuffer(nil)
	at := table.NewEmbeddedTable(buf)
	at.SetHeader([]string{" "})
	at.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, s := range subnets {
		at.Append([]string{s.Name, s.IPv4Block, s.ID.String()})
	}
	at.Render()

	return buf.String()
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "-"
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, labels[k]))
	}

	return strings.Join(pairs, ", ")
}
