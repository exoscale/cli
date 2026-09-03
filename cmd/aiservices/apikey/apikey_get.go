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

type AIAPIKeyShowOutput struct {
	ID        v3.UUID `json:"id"`
	Name      string  `json:"name"`
	Scope     string  `json:"scope"`
	CreatedAt string  `json:"created_at" outputLabel:"Created At"`
	UpdatedAt string  `json:"updated_at" outputLabel:"Updated At"`
}

func (o *AIAPIKeyShowOutput) ToJSON()  { output.JSON(o) }
func (o *AIAPIKeyShowOutput) ToText()  { output.Text(o) }
func (o *AIAPIKeyShowOutput) ToTable() { output.Table(o) }

type AIAPIKeyGetCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	ID string `cli-arg:"#" cli-usage:"AI API key ID"`
}

func (c *AIAPIKeyGetCmd) CmdAliases() []string { return exocmd.GShowAlias }
func (c *AIAPIKeyGetCmd) CmdShort() string     { return "Show AI API key details" }
func (c *AIAPIKeyGetCmd) CmdLong() string {
	return "This command shows details of an AI API key by ID."
}
func (c *AIAPIKeyGetCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyGetCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	if c.ID == "" {
		return fmt.Errorf("ID is required")
	}

	id, err := v3.ParseUUID(c.ID)
	if err != nil {
		return fmt.Errorf("invalid AI API key ID: %w", err)
	}

	resp, err := client.GetAIAPIKey(ctx, id)
	if err != nil {
		return err
	}

	out := &AIAPIKeyShowOutput{
		ID:        resp.ID,
		Name:      resp.Name,
		Scope:     resp.Scope,
		CreatedAt: resp.CreatedAT.Format(time.RFC3339),
		UpdatedAt: resp.UpdatedAT.Format(time.RFC3339),
	}

	return c.OutputFunc(out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyGetCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
