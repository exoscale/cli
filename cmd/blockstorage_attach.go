package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
)

type blockstorageAttachCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
}

func (c *blockstorageAttachCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageAttachCmd) cmdShort() string { return "Attach a Block Storage Volume" }

func (c *blockstorageAttachCmd) cmdLong() string {
	return fmt.Sprintf(`This command attachs Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageAttachCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageAttachCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageAttachCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
