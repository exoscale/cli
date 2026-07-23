package vpc

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type vpcSubnetInstanceOutput struct {
	ID   v3.UUID `json:"id"`
	IPv4 string  `json:"ipv4"`
}

type vpcSubnetShowOutput struct {
	ID            v3.UUID                   `json:"id"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	Zone          v3.ZoneName               `json:"zone"`
	CreatedAt     string                    `json:"created_at"`
	AddressFamily string                    `json:"address_family"`
	AddressSpace  string                    `json:"address_space"`
	IPv4Block     string                    `json:"ipv4_block"`
	Labels        map[string]string         `json:"labels"`
	Instances     []vpcSubnetInstanceOutput `json:"instances"`
}

func (o *vpcSubnetShowOutput) ToJSON() { output.JSON(o) }
func (o *vpcSubnetShowOutput) ToText() { output.Text(o) }
func (o *vpcSubnetShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"VPC Subnet"})
	defer t.Render()

	t.Append([]string{"ID", o.ID.String()})
	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Description", o.Description})
	t.Append([]string{"Zone", string(o.Zone)})
	t.Append([]string{"Created At", o.CreatedAt})
	t.Append([]string{"Address Family", o.AddressFamily})
	t.Append([]string{"Address Space", o.AddressSpace})
	t.Append([]string{"IPv4 Block", o.IPv4Block})
	t.Append([]string{"Labels", formatLabels(o.Labels)})
	t.Append([]string{"Instances", formatSubnetInstances(o.Instances)})
}

type vpcSubnetShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	VPC    string `cli-arg:"#" cli-usage:"VPC-NAME|ID"`
	Subnet string `cli-arg:"#" cli-usage:"SUBNET-NAME|ID"`

	Zone v3.ZoneName `cli-short:"z" cli-usage:"VPC zone"`
}

func (c *vpcSubnetShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *vpcSubnetShowCmd) CmdShort() string { return "Show a VPC Subnet details" }

func (c *vpcSubnetShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a VPC Subnet details.

Supported output template annotations for Subnet: %s

Supported output template annotations for Subnet instances: %s`,
		strings.Join(output.TemplateAnnotations(&vpcSubnetShowOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&vpcSubnetInstanceOutput{}), ", "))
}

func (c *vpcSubnetShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *vpcSubnetShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	vpcEntry, err := FindVPC(ctx, client, c.VPC)
	if err != nil {
		return err
	}

	subnetEntry, err := FindSubnet(ctx, client, vpcEntry.ID, c.Subnet)
	if err != nil {
		return err
	}

	subnet, err := client.GetSubnet(ctx, vpcEntry.ID, subnetEntry.ID)
	if err != nil {
		return err
	}

	out := vpcSubnetShowOutput{
		ID:            subnet.ID,
		Name:          subnet.Name,
		Description:   subnet.Description,
		Zone:          c.Zone,
		CreatedAt:     subnet.CreatedAT.String(),
		AddressFamily: string(subnet.Addressfamily),
		AddressSpace:  string(subnet.AddressSpace),
		IPv4Block:     subnet.Ipv4Block,
		Labels:        subnet.Labels,
		Instances:     []vpcSubnetInstanceOutput{},
	}

	for _, i := range subnet.Instances {
		ipv4 := ""
		if i.Ipv4 != nil {
			ipv4 = i.Ipv4.String()
		}
		out.Instances = append(out.Instances, vpcSubnetInstanceOutput{
			ID:   i.ID,
			IPv4: ipv4,
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(vpcSubnetCmd, &vpcSubnetShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}

func formatSubnetInstances(instances []vpcSubnetInstanceOutput) string {
	if len(instances) == 0 {
		return "-"
	}

	buf := bytes.NewBuffer(nil)
	at := table.NewEmbeddedTable(buf)
	at.SetHeader([]string{" "})
	at.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, i := range instances {
		at.Append([]string{i.ID.String(), i.IPv4})
	}
	at.Render()

	return buf.String()
}
