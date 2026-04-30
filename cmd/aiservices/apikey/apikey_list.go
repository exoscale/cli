package apikey

import (
	"fmt"
	"strings"
	"time"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
	"github.com/spf13/cobra"
)

type AIAPIKeyListItemOutput struct {
	ID        v3.UUID `json:"id"`
	Name      string  `json:"name"`
	Scope     string  `json:"scope"`
	CreatedAt string  `json:"created_at" outputLabel:"Created At"`
}

type AIAPIKeyListOutput []AIAPIKeyListItemOutput

func (o *AIAPIKeyListOutput) ToJSON()  { output.JSON(o) }
func (o *AIAPIKeyListOutput) ToText()  { output.Text(o) }
func (o *AIAPIKeyListOutput) ToTable() { output.Table(o) }

type AIAPIKeyListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *AIAPIKeyListCmd) CmdAliases() []string { return exocmd.GListAlias }
func (c *AIAPIKeyListCmd) CmdShort() string     { return "List AI API keys" }
func (c *AIAPIKeyListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists AI API keys.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&AIAPIKeyListOutput{}), ", "))
}
func (c *AIAPIKeyListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}
func (c *AIAPIKeyListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	resp, err := client.ListAIAPIKeys(ctx)
	if err != nil {
		return err
	}

	out := make(AIAPIKeyListOutput, 0, len(resp.AIAPIKeys))
	for _, key := range resp.AIAPIKeys {
		out = append(out, AIAPIKeyListItemOutput{
			ID:        key.ID,
			Name:      key.Name,
			Scope:     key.Scope,
			CreatedAt: key.CreatedAT.Format(time.RFC3339),
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(Cmd, &AIAPIKeyListCmd{CliCommandSettings: exocmd.DefaultCLICmdSettings()}))
}
