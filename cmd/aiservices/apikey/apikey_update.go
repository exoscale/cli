package apikey

import (
	"fmt"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyUpdateOutput struct {
	ID             v3.UUID  `json:"id"`
	Name           string   `json:"name"`
	Models         []string `json:"models"`
	Deployments    []string `json:"deployments"`
	AllModels      bool     `json:"all_models"`
	AllDeployments bool     `json:"all_deployments"`
	UpdatedAt      string   `json:"updated_at"`
}

func (o *AIAPIKeyUpdateOutput) ToJSON()  { output.JSON(o) }
func (o *AIAPIKeyUpdateOutput) ToText()  { output.Text(o) }
func (o *AIAPIKeyUpdateOutput) ToTable() { output.Table(o) }

type AIAPIKeyUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Key            string      `cli-arg:"#" cli-usage:"ID or NAME"`
	Deployments    []string    `cli-flag:"deployment" cli-usage:"Deployment ID the API key can access (repeatable)"`
	Models         []string    `cli-flag:"model" cli-usage:"Public model name the API key can access (repeatable)"`
	AllModels      bool        `cli-usage:"Grant access to all public models; set --all-models=false to revoke"`
	AllDeployments bool        `cli-usage:"Grant access to all deployments; set --all-deployments=false to revoke"`
	Zone           v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *AIAPIKeyUpdateCmd) CmdAliases() []string { return exocmd.GUpdateAlias }
func (c *AIAPIKeyUpdateCmd) CmdShort() string     { return "Update AI API key" }
func (c *AIAPIKeyUpdateCmd) CmdLong() string {
	return "This command updates the models and/or deployments accessible by an AI API key. Omitted properties are left unchanged."
}
func (c *AIAPIKeyUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.Key == "" {
		return fmt.Errorf("ID or NAME is required")
	}

	allModelsChanged := cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AllModels))
	allDeploymentsChanged := cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.AllDeployments))

	if !allModelsChanged && !allDeploymentsChanged && len(c.Models) == 0 && len(c.Deployments) == 0 {
		return fmt.Errorf("at least one of --deployment, --model, --all-models or --all-deployments is required")
	}
	if allModelsChanged && len(c.Models) > 0 {
		return conflictingFlagsErr("model", "all-models")
	}
	if allDeploymentsChanged && len(c.Deployments) > 0 {
		return conflictingFlagsErr("deployment", "all-deployments")
	}

	req := v3.UpdateAIAPIKeyRequest{}
	if allModelsChanged {
		req.AllModels = &c.AllModels
	} else if len(c.Models) > 0 {
		models := v3.AIAPIKeyModels(c.Models)
		req.Models = &models
	}
	if allDeploymentsChanged {
		req.AllDeployments = &c.AllDeployments
	} else if len(c.Deployments) > 0 {
		deployments, err := deploymentsToRefs(c.Deployments)
		if err != nil {
			return err
		}
		req.Deployments = &deployments
	}

	list, err := client.ListAIAPIKeys(ctx)
	if err != nil {
		return err
	}
	entry, err := list.FindListAIAPIKeysResponseEntry(c.Key)
	if err != nil {
		return err
	}

	resp, err := client.UpdateAIAPIKey(ctx, entry.ID, req)
	if err != nil {
		return err
	}

	out := &AIAPIKeyUpdateOutput{
		ID:             resp.ID,
		Name:           resp.Name,
		Models:         modelsToStrings(resp.Models),
		Deployments:    deploymentsToStrings(resp.Deployments),
		AllModels:      derefBool(resp.AllModels),
		AllDeployments: derefBool(resp.AllDeployments),
		UpdatedAt:      resp.UpdatedAT.Format(time.RFC3339),
	}

	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyUpdateCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
