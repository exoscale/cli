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

type blockstorageDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#" cli-usage:"NAME"`
	Zone string `cli-short:"z" cli-usage:"block storage volume zone"`
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
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	client := globalstate.EgoscaleV3Client
	TODO := context.TODO()

	resp, err := client.ListBlockStorageVolumes(TODO)
	if err != nil {
		return err
	}
	volume, err := resp.FindBlockStorageVolume(c.Name)
	if err != nil {
		return err
	}

	op, err := client.DeleteBlockStorageVolume(TODO, volume.ID)
	if err != nil {
		return err
	}
	_, err = client.Wait(TODO, op, v3.OperationStateSuccess)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
