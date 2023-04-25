package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceSnapshotExportOutput struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

func (o *instanceSnapshotExportOutput) ToJSON()  { output.JSON(o) }
func (o *instanceSnapshotExportOutput) ToText()  { output.Text(o) }
func (o *instanceSnapshotExportOutput) ToTable() { output.Table(o) }

type instanceSnapshotExportCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"export"`

	ID string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotExportCmd) cmdAliases() []string { return nil }

func (c *instanceSnapshotExportCmd) cmdShort() string {
	return "Export a Compute instance snapshot"
}

func (c *instanceSnapshotExportCmd) cmdLong() string {
	return fmt.Sprintf(`This command exports a Compute instance snapshot.
	
Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceSnapshotExportOutput{}), ", "))
}

func (c *instanceSnapshotExportCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotExportCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	snapshot, err := globalstate.GlobalEgoscaleClient.GetSnapshot(ctx, c.Zone, c.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	var snapshotExport *egoscale.SnapshotExport
	decorateAsyncOperation(fmt.Sprintf("Exporting snapshot %s...", c.ID), func() {
		snapshotExport, err = globalstate.GlobalEgoscaleClient.ExportSnapshot(ctx, c.Zone, snapshot)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.outputFunc(
			&instanceSnapshotExportOutput{
				URL:      *snapshotExport.PresignedURL,
				Checksum: *snapshotExport.MD5sum,
			},
			nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceSnapshotCmd, &instanceSnapshotExportCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
