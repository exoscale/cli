package sks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type sksActiveNodepoolTemplateOutput struct {
	KubeVersion string  `json:"kube_version"`
	Variant     string  `json:"variant"`
	TemplateID  v3.UUID `json:"template_id"`
	Template    string  `json:"template"`
}

type sksActiveNodepoolTemplatesOutput []sksActiveNodepoolTemplateOutput

func (o *sksActiveNodepoolTemplatesOutput) ToJSON()  { output.JSON(o) }
func (o *sksActiveNodepoolTemplatesOutput) ToText()  { output.Text(o) }
func (o *sksActiveNodepoolTemplatesOutput) ToTable() { output.Table(o) }

type sksActiveNodepoolTemplatesCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"active-nodepool-templates"`

	KubeVersion string      `cli-arg:"#" cli-usage:"KUBERNETES-VERSION"`
	Variant     string      `cli-flag:"variant" cli-usage:"nodepool template variant to resolve (standard|nvidia)"`
	Zone        v3.ZoneName `cli-short:"z" cli-usage:"zone to query in"`
}

func (c *sksActiveNodepoolTemplatesCmd) CmdAliases() []string {
	return []string{"active-nodepool-template"}
}

func (c *sksActiveNodepoolTemplatesCmd) CmdShort() string {
	return "Find active SKS nodepool templates for a Kubernetes version"
}

func (c *sksActiveNodepoolTemplatesCmd) CmdLong() string {
	return fmt.Sprintf(`This command finds active SKS nodepool templates for a given Kubernetes version.

By default, both "standard" and "nvidia" variants are queried.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sksActiveNodepoolTemplateOutput{}), ", "))
}

func (c *sksActiveNodepoolTemplatesCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *sksActiveNodepoolTemplatesCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	var variants []v3.GetActiveNodepoolTemplateVariant
	switch c.Variant {
	case "":
		variants = []v3.GetActiveNodepoolTemplateVariant{
			v3.GetActiveNodepoolTemplateVariantStandard,
			v3.GetActiveNodepoolTemplateVariantNvidia,
		}
	case string(v3.GetActiveNodepoolTemplateVariantStandard):
		variants = []v3.GetActiveNodepoolTemplateVariant{v3.GetActiveNodepoolTemplateVariantStandard}
	case string(v3.GetActiveNodepoolTemplateVariantNvidia):
		variants = []v3.GetActiveNodepoolTemplateVariant{v3.GetActiveNodepoolTemplateVariantNvidia}
	default:
		return errors.New(`invalid variant, must be one of: "standard", "nvidia"`)
	}

	out := make(sksActiveNodepoolTemplatesOutput, 0, len(variants))
	for _, variant := range variants {
		activeTemplate, err := client.GetActiveNodepoolTemplate(ctx, c.KubeVersion, variant)
		if err != nil {
			return fmt.Errorf("error retrieving active %q nodepool template: %w", variant, err)
		}

		if activeTemplate.ActiveTemplate == "" {
			return fmt.Errorf("no active template returned for variant %q", variant)
		}

		template, err := client.GetTemplate(ctx, activeTemplate.ActiveTemplate)
		if err != nil {
			return fmt.Errorf("error retrieving template details for %q: %w", activeTemplate.ActiveTemplate, err)
		}

		out = append(out, sksActiveNodepoolTemplateOutput{
			KubeVersion: c.KubeVersion,
			Variant:     string(variant),
			TemplateID:  template.ID,
			Template:    template.Name,
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(sksCmd, &sksActiveNodepoolTemplatesCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
