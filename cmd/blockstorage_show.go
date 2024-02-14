package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
)

type blockstorageShowOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (o *blockstorageShowOutput) Type() string { return "Block Storage Volume" }
func (o *blockstorageShowOutput) ToJSON()      { output.JSON(o) }
func (o *blockstorageShowOutput) ToText()      { output.Text(o) }
func (o *blockstorageShowOutput) ToTable()     { output.Table(o) }

type blockstorageShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone     string `cli-short:"z" cli-usage:"block storage volume zone"`
}

func (c *blockstorageShowCmd) cmdAliases() []string { return gShowAlias }

func (c *blockstorageShowCmd) cmdShort() string { return "Show a Block Storage Volume details" }

func (c *blockstorageShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows a Block Storage Volume details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "))
}

func (c *blockstorageShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageShowCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
