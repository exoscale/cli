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
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceTypeListItemOutput struct {
	ID         string `json:"id"`
	Family     string `json:"family"`
	Size       string `json:"name"`
	CPUs       int64  `json:"cpus"`
	Memory     int64  `json:"memory"`
	Authorized bool   `json:"authorized"`
}

type instanceTypeListOutput struct {
	data    []instanceTypeListItemOutput
	verbose bool
}

func (o *instanceTypeListOutput) ToJSON() { output.JSON(o.data) }
func (o *instanceTypeListOutput) ToText() { output.Text(o.data) }
func (o *instanceTypeListOutput) ToTable() {
	header := []string{"ID", "Family", "Size"}
	if o.verbose {
		header = append(header, "# CPUs", "Memory", "Authorized")
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader(header)
	defer t.Render()

	for _, p := range o.data {
		cols := []string{p.ID, p.Family, p.Size}

		if o.verbose {
			cols = append(
				cols,
				fmt.Sprint(p.CPUs),
				humanize.Bytes(uint64(p.Memory)),
				fmt.Sprint(p.Authorized),
			)
		}

		t.Append(cols)
	}
}

type instanceTypeListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Verbose bool `cli-short:"v" cli-help:"show additional information about Compute instance types"`
}

func (c *instanceTypeListCmd) CmdAliases() []string { return nil }

func (c *instanceTypeListCmd) CmdShort() string { return "List Compute instance types" }

func (c *instanceTypeListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists available Compute instance types.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceTypeListItemOutput{}), ", "))
}

func (c *instanceTypeListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTypeListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(exocmd.DefaultZone))
	if err != nil {
		return err
	}
	instanceTypes, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return err
	}

	out := instanceTypeListOutput{
		data:    make([]instanceTypeListItemOutput, 0),
		verbose: c.Verbose,
	}

	for _, t := range instanceTypes.InstanceTypes {
		out.data = append(out.data, instanceTypeListItemOutput{
			ID:         t.ID.String(),
			Family:     string(t.Family),
			Size:       string(t.Size),
			Memory:     t.Memory,
			CPUs:       t.Cpus,
			Authorized: *t.Authorized,
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceTypeCmd, &instanceTypeListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
