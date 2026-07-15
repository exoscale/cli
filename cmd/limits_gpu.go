package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type GpuLimitOutput struct {
	Resource string `json:"resource"`
	Used     int64  `json:"used"`
	Max      int64  `json:"max"`
}

type GpuLimitsOutput []GpuLimitOutput

func (o *GpuLimitsOutput) ToJSON()  { output.JSON(o) }
func (o *GpuLimitsOutput) ToText()  { output.Text(o) }
func (o *GpuLimitsOutput) ToTable() { output.Table(o) }

var gpuResourceLabels = map[string]string{
	gpu2:          "GPU - GPU2",
	gpu3:          "GPU - GPU3",
	gpua30:        "GPU - A30",
	gpu3080ti:     "GPU - 3080 Ti",
	gpua5000:      "GPU - A5000",
	gpurtx6000pro: "GPU - RTX 6000 Pro",
}

type LimitsGpuCmd struct {
	CliCommandSettings `cli-cmd:"-"`
	_                  bool        `cli-cmd:"gpu"`
	Zone               v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *LimitsGpuCmd) CmdAliases() []string { return nil }
func (c *LimitsGpuCmd) CmdShort() string     { return "Show per-family GPU limits" }
func (c *LimitsGpuCmd) CmdLong() string {
	return strings.Join([]string{
		"Show per-family GPU quotas (A5000, A30, 3080 Ti, RTX 6000 Pro, GPU2, GPU3).",
		"",
		fmt.Sprintf("Supported output template annotations: %s",
			strings.Join(output.TemplateAnnotations(&GpuLimitsOutput{}), ", ")),
	}, "\n")
}

func (c *LimitsGpuCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *LimitsGpuCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	quotas, err := client.ListQuotas(ctx)
	if err != nil {
		return err
	}

	out := GpuLimitsOutput{}
	for _, quota := range quotas.Quotas {
		if label, ok := gpuResourceLabels[quota.Resource]; ok {
			out = append(out, GpuLimitOutput{
				Resource: label,
				Used:     quota.Usage,
				Max:      quota.Limit,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Resource < out[j].Resource
	})

	return utils.PrintOutput(&out, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(limitsCmd, &LimitsGpuCmd{CliCommandSettings: DefaultCLICmdSettings()}))
}
