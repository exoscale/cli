package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceTypeShowOutput struct {
	ID         string `json:"id"`
	Family     string `json:"family"`
	Size       string `json:"name"`
	Memory     int64  `json:"memory"`
	CPUs       int64  `json:"cpus"`
	GPUs       int64  `json:"gpus"`
	Authorized bool   `json:"authorized"`
}

func (o *instanceTypeShowOutput) ToJSON() { output.JSON(o) }
func (o *instanceTypeShowOutput) ToText() { output.Text(o) }
func (o *instanceTypeShowOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"Instance Type"})
	defer t.Render()

	t.Append([]string{"ID", o.ID})
	t.Append([]string{"Family", o.Family})
	t.Append([]string{"Size", o.Size})
	t.Append([]string{"Memory", humanize.IBytes(uint64(o.Memory))})
	t.Append([]string{"# CPUs", fmt.Sprint(o.CPUs)})

	if o.GPUs > 0 {
		t.Append([]string{"# GPUs", fmt.Sprint(o.GPUs)})
	}

	t.Append([]string{"Authorized", fmt.Sprint(o.Authorized)})
}

type instanceTypeShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type string `cli-arg:"#" cli-usage:"[FAMILY.]SIZE"`
}

func (c *instanceTypeShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *instanceTypeShowCmd) CmdShort() string {
	return "Show a Compute instance type details"
}

func (c *instanceTypeShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance type details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceTypeShowOutput{}), ", "))
}

func (c *instanceTypeShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTypeShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(exocmd.DefaultZone))
	if err != nil {
		return err
	}
	instanceTypes, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return err
	}

	t, err := instanceTypes.FindInstanceTypeByIdOrFamilyAndSize(c.Type)
	if err != nil {
		return err
	}
	return c.OutputFunc(&instanceTypeShowOutput{
		ID:         t.ID.String(),
		Family:     string(t.Family),
		Size:       string(t.Size),
		Memory:     t.Memory,
		CPUs:       t.Cpus,
		GPUs:       t.Gpus,
		Authorized: utils.DefaultBool(t.Authorized, false),
	}, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceTypeCmd, &instanceTypeShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
