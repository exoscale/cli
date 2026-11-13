package bucket

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
)

const (
	versioningStatus  = "status"
	versioningEnable  = "enable"
	versioningSuspend = "suspend"
)

type versioningCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"versioning"`

	Operation string `cli-arg:"#" cli-usage:"{status,enable,suspend}"`
	Bucket    string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	Zone string `cli-short:"z" cli-usage:"bucket zone"`
}

func (c *versioningCmd) CmdAliases() []string { return []string{"v"} }
func (c *versioningCmd) CmdShort() string {
	return "Manage the Object Versioning setting of a Storage Bucket"
}
func (c *versioningCmd) CmdLong() string {
	return "Manage the Object Versioning setting of a Storage Bucket"
}

func (c *versioningCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if !slices.Contains([]string{versioningStatus, versioningEnable, versioningSuspend}, args[0]) {
		clicmd.CmdExitOnUsageError(cmd, "invalid operation")
	}

	clicmd.CmdSetZoneFlagFromDefault(cmd)
	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"zone"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *versioningCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptWithZone(c.Zone),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	switch c.Operation {
	case versioningStatus:
		return c.OutputFunc(storage.BucketVersioningStatus(cmd.Context(), bucket))
	case versioningEnable:
		return storage.EnableBucketVersioning(cmd.Context(), bucket)
	case versioningSuspend:
		return storage.SuspendBucketVersioning(cmd.Context(), bucket)
	}

	return fmt.Errorf("invalid operation")
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &versioningCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
