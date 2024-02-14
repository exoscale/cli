package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
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
	client := globalstate.EgoscaleV3Client
	TODO := context.TODO()

	op, err := client.CreateBlockStorageVolume(TODO, v3.CreateBlockStorageVolumeRequest{})
	if err != nil {
		return err
	}
	op, err = client.Wait(TODO, op, v3.OperationStateSuccess)
	if err != nil {
		return err
	}

	bs, err := client.GetBlockStorageVolume(TODO, op.Reference.ID)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&blockstorageShowCmd{
			Name: bs.Name,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
