package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Type string `cli-arg:"#" cli-usage:"[FAMILY.]SIZE"`
}

func (c *instanceTypeShowCmd) cmdAliases() []string { return gShowAlias }

func (c *instanceTypeShowCmd) cmdShort() string {
	return "Show a Compute instance type details"
}

func (c *instanceTypeShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Compute instance type details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceTypeShowOutput{}), ", "))
}

func (c *instanceTypeShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceTypeShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	t, err := globalstate.EgoscaleClient.FindInstanceType(ctx, account.CurrentAccount.DefaultZone, c.Type)
	if err != nil {
		return err
	}

	return c.outputFunc(&instanceTypeShowOutput{
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
		Authorized: *t.Authorized,
	}, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceTypeCmd, &instanceTypeShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
