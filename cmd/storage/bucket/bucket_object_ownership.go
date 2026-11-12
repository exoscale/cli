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
	objectOwnershipStatus               = "status"
	objectOwnershipObjectWriter         = "object-writer"
	objectOwnershipBucketOwnerEnforced  = "bucket-owner-enforced"
	objectOwnershipBucketOwnerPreferred = "bucket-owner-preferred"
)

type objectOwnershipCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"object-ownership"`

	Operation string `cli-arg:"#" cli-usage:"{status,object-writer,bucket-owner-enforced,bucket-owner-preferred}"`
	Bucket    string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	Zone string `cli-short:"z" cli-usage:"bucket zone"`
}

func (c *objectOwnershipCmd) CmdAliases() []string { return []string{"oo"} }
func (c *objectOwnershipCmd) CmdShort() string {
	return "Manage the Object Ownership setting of a Storage Bucket"
}
func (c *objectOwnershipCmd) CmdLong() string {
	return "Manage the Object Ownership setting of a Storage Bucket"
}

func (c *objectOwnershipCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if !slices.Contains([]string{
		objectOwnershipStatus,
		objectOwnershipObjectWriter,
		objectOwnershipBucketOwnerEnforced,
		objectOwnershipBucketOwnerPreferred,
	}, args[0]) {
		clicmd.CmdExitOnUsageError(cmd, "invalid operation")
	}

	clicmd.CmdSetZoneFlagFromDefault(cmd)
	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"zone"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *objectOwnershipCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	bucket := strings.TrimPrefix(c.Bucket, sos.BucketPrefix)

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptWithZone(c.Zone),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	switch c.Operation {
	case objectOwnershipStatus:
		return c.OutputFunc(storage.GetBucketObjectOwnershipInfo(cmd.Context(), bucket))
	case objectOwnershipObjectWriter:
		return storage.SetBucketObjectOwnership(cmd.Context(), bucket, sos.ObjectOwnershipObjectWriter)
	case objectOwnershipBucketOwnerPreferred:
		return storage.SetBucketObjectOwnership(cmd.Context(), bucket, sos.ObjectOwnershipBucketOwnerPreferred)
	case objectOwnershipBucketOwnerEnforced:
		return storage.SetBucketObjectOwnership(cmd.Context(), bucket, sos.ObjectOwnershipBucketOwnerEnforced)
	}

	return fmt.Errorf("invalid operation")
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &objectOwnershipCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
