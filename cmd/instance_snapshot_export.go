package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instanceSnapshotExportOutput struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

func (o *instanceSnapshotExportOutput) ToJSON()  { output.JSON(o) }
func (o *instanceSnapshotExportOutput) ToText()  { output.Text(o) }
func (o *instanceSnapshotExportOutput) ToTable() { output.Table(o) }

type instanceSnapshotExportCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"export"`

	ID string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotExportCmd) CmdAliases() []string { return nil }

func (c *instanceSnapshotExportCmd) CmdShort() string {
	return "Export a Compute instance snapshot"
}

func (c *instanceSnapshotExportCmd) CmdLong() string {
	return fmt.Sprintf(`This command exports a Compute instance snapshot.
	
Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceSnapshotExportOutput{}), ", "))
}

func (c *instanceSnapshotExportCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotExportCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	snapshot, err := globalstate.EgoscaleClient.GetSnapshot(ctx, c.Zone, c.ID)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	var snapshotExport *egoscale.SnapshotExport
	decorateAsyncOperation(fmt.Sprintf("Exporting snapshot %s...", c.ID), func() {
		snapshotExport, err = globalstate.EgoscaleClient.ExportSnapshot(ctx, c.Zone, snapshot)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc(
			&instanceSnapshotExportOutput{
				URL:      *snapshotExport.PresignedURL,
				Checksum: *snapshotExport.MD5sum,
			},
			nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotExportCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
