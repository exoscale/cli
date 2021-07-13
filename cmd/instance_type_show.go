package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeInstanceTypeShowOutput struct {
	ID     string `json:"id"`
	Family string `json:"family"`
	Size   string `json:"name"`
	Memory int64  `json:"memory"`
	CPUs   int64  `json:"cpus"`
	GPUs   int64  `json:"gpus"`
}

func (o *computeInstanceTypeShowOutput) toJSON() { outputJSON(o) }
func (o *computeInstanceTypeShowOutput) toText() { outputText(o) }
func (o *computeInstanceTypeShowOutput) toTable() {
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
}

type computeInstanceTypeShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type string `cli-arg:"#" cli-usage:"[FAMILY.]SIZE"`
}

func (c *computeInstanceTypeShowCmd) cmdAliases() []string { return gShowAlias }

func (c *computeInstanceTypeShowCmd) cmdShort() string {
	return "Show a Compute instance type details"
}

func (c *computeInstanceTypeShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance type details.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&computeInstanceTypeShowOutput{}), ", "))
}

func (c *computeInstanceTypeShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeInstanceTypeShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	t, err := cs.FindInstanceType(ctx, gCurrentAccount.DefaultZone, c.Type)
	if err != nil {
		return err
	}

	return output(&computeInstanceTypeShowOutput{
		ID:     *t.ID,
		Family: *t.Family,
		Size:   *t.Size,
		Memory: *t.Memory,
		CPUs:   *t.CPUs,
		GPUs: func() (v int64) {
			if t.GPUs != nil {
				v = *t.GPUs
			}
			return
		}(),
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceTypeCmd, &computeInstanceTypeShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
