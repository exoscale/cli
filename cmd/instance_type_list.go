package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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

func (o *instanceTypeListOutput) toJSON() { output.JSON(o.data) }
func (o *instanceTypeListOutput) toText() { output.Text(o.data) }
func (o *instanceTypeListOutput) toTable() {
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Verbose bool `cli-short:"v" cli-help:"show additional information about Compute instance types"`
}

func (c *instanceTypeListCmd) cmdAliases() []string { return nil }

func (c *instanceTypeListCmd) cmdShort() string { return "List Compute instance types" }

func (c *instanceTypeListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists available Compute instance types.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceTypeListItemOutput{}), ", "))
}

func (c *instanceTypeListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTypeListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	instanceTypes, err := cs.ListInstanceTypes(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := instanceTypeListOutput{
		data:    make([]instanceTypeListItemOutput, 0),
		verbose: c.Verbose,
	}

	for _, t := range instanceTypes {
		out.data = append(out.data, instanceTypeListItemOutput{
			ID:         *t.ID,
			Family:     *t.Family,
			Size:       *t.Size,
			Memory:     *t.Memory,
			CPUs:       *t.CPUs,
			Authorized: *t.Authorized,
		})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceTypeCmd, &instanceTypeListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
