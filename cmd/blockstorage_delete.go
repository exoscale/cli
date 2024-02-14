package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
)

type blockstorageDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
}

func (c *blockstorageDeleteCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageDeleteCmd) cmdShort() string { return "Delete a Block Storage Volume" }

func (c *blockstorageDeleteCmd) cmdLong() string {
	return fmt.Sprintf(`This command deletes Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	cmdSetTemplateFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
