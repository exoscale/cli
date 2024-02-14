package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
)

type blockstorageListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
}

func (c *blockstorageListCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageListCmd) cmdShort() string { return "List Block Storage Volumes" }

func (c *blockstorageListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Block Storage Volumes.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
