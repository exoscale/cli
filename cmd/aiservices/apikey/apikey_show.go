package apikey

import (
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyShowOutput struct {
	ID             v3.UUID  `json:"id"`
	Name           string   `json:"name"`
	Models         []string `json:"models"`
	Deployments    []string `json:"deployments"`
	AllModels      bool     `json:"all_models"`
	AllDeployments bool     `json:"all_deployments"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

func (o *AIAPIKeyShowOutput) ToJSON()  { output.JSON(o) }
func (o *AIAPIKeyShowOutput) ToText()  { output.Text(o) }
func (o *AIAPIKeyShowOutput) ToTable() { output.Table(o) }

type AIAPIKeyShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Key  string      `cli-arg:"#"  cli-usage:"ID or NAME"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"zone"`
}

func (c *AIAPIKeyShowCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *AIAPIKeyShowCmd) CmdShort() string     { return "Show AI API key details" }
func (c *AIAPIKeyShowCmd) CmdLong() string {
	return "This command shows details of an AI API key by ID or name."
}
func (c *AIAPIKeyShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	list, err := client.ListAIAPIKeys(ctx)
	if err != nil {
		return err
	}
	entry, err := list.FindListAIAPIKeysResponseEntry(c.Key)
	if err != nil {
		return err
	}

	resp, err := client.GetAIAPIKey(ctx, entry.ID)
	if err != nil {
		return err
	}

	out := &AIAPIKeyShowOutput{
		ID:             resp.ID,
		Name:           resp.Name,
		Models:         modelsToStrings(resp.Models),
		Deployments:    deploymentsToStrings(resp.Deployments),
		AllModels:      derefBool(resp.AllModels),
		AllDeployments: derefBool(resp.AllDeployments),
		CreatedAt:      resp.CreatedAT.Format(time.RFC3339),
		UpdatedAt:      resp.UpdatedAT.Format(time.RFC3339),
	}

	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyShowCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
