package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
)

type blockstorageDetachCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
}

func (c *blockstorageDetachCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageDetachCmd) cmdShort() string { return "Detach a Block Storage Volume" }

func (c *blockstorageDetachCmd) cmdLong() string {
	return fmt.Sprintf(`This command detaches Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageDetachCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageDetachCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageDetachCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
