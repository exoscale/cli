package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
)

type blockstorageCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
}

func (c *blockstorageCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageCreateCmd) cmdShort() string { return "Create a Block Storage Volume" }

func (c *blockstorageCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
