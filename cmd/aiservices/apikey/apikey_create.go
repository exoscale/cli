package apikey

import (
	"fmt"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyCreateOutput struct {
	ID             v3.UUID  `json:"id"`
	Name           string   `json:"name"`
	Value          string   `json:"value"`
	Models         []string `json:"models"`
	Deployments    []string `json:"deployments"`
	AllModels      bool     `json:"all_models"`
	AllDeployments bool     `json:"all_deployments"`
}

func (o *AIAPIKeyCreateOutput) ToJSON()  { output.JSON(o) }
func (o *AIAPIKeyCreateOutput) ToText()  { output.Text(o) }
func (o *AIAPIKeyCreateOutput) ToTable() { output.Table(o) }

type AIAPIKeyCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name           string      `cli-arg:"#" cli-usage:"NAME"`
	Deployments    []string    `cli-flag:"deployment" cli-usage:"Deployment ID the API key can access (repeatable)"`
	Models         []string    `cli-flag:"model" cli-usage:"Public model name the API key can access (repeatable)"`
	AllModels      bool        `cli-usage:"Grant access to all public models"`
	AllDeployments bool        `cli-usage:"Grant access to all deployments"`
	Zone           v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *AIAPIKeyCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }
func (c *AIAPIKeyCreateCmd) CmdShort() string     { return "Create AI API key" }
func (c *AIAPIKeyCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates an AI API key.

The API key value is only returned once, at creation time, so store it securely.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&AIAPIKeyCreateOutput{}), ", "))
}
func (c *AIAPIKeyCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("NAME is required")
	}

	if c.AllModels && len(c.Models) > 0 {
		return conflictingFlagsErr("model", "all-models")
	}
	if c.AllDeployments && len(c.Deployments) > 0 {
		return conflictingFlagsErr("deployment", "all-deployments")
	}

	req := v3.CreateAIAPIKeyRequest{Name: c.Name}
	if c.AllModels {
		req.AllModels = new(true)
	} else if len(c.Models) > 0 {
		models := v3.AIAPIKeyModels(c.Models)
		req.Models = &models
	}
	if c.AllDeployments {
		req.AllDeployments = new(true)
	} else if len(c.Deployments) > 0 {
		deployments, err := deploymentsToRefs(c.Deployments)
		if err != nil {
			return err
		}
		req.Deployments = &deployments
	}

	resp, err := client.CreateAIAPIKey(ctx, req)
	if err != nil {
		return err
	}

	out := &AIAPIKeyCreateOutput{
		ID:             resp.ID,
		Name:           resp.Name,
		Value:          resp.Value,
		Models:         modelsToStrings(resp.Models),
		Deployments:    deploymentsToStrings(resp.Deployments),
		AllModels:      derefBool(resp.AllModels),
		AllDeployments: derefBool(resp.AllDeployments),
	}

	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyCreateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
