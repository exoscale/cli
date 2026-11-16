package storage

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type storageMbCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"mb"`

	Bucket string `cli-arg:"#" cli-usage:"sos://BUCKET"`

	ACL  string `cli-usage:"canned ACL to set on bucket (private|public-read|public-read-write|authenticated-read), see https://community.exoscale.com/product/storage/object-storage/how-to/acl/"`
	Zone string `cli-short:"z" cli-usage:"bucket zone"`
}

func (c *storageMbCmd) CmdAliases() []string { return []string{"create"} }
func (c *storageMbCmd) CmdShort() string     { return "Create a new bucket" }
func (c *storageMbCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a new bucket.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowBucketOutput{}), ", "))
}

func (c *storageMbCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	target := args[0]
	args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
	args[0] = strings.TrimRight(args[0], "/")

	if args[0] == "" {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", target))
	}

	clicmd.CmdSetZoneFlagFromDefault(cmd)

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"zone"}); err != nil {
		return err
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if c.ACL != "" && !slices.Contains(sos.BucketCannedACLToStrings(), c.ACL) {
		clicmd.CmdExitOnUsageError(cmd, fmt.Sprintf("invalid canned ACL %q, supported values are: %s",
			c.ACL, strings.Join(sos.BucketCannedACLToStrings(), ", ")))
	}

	return nil
}

func (c *storageMbCmd) CmdRun(_ *cobra.Command, _ []string) error {
	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptWithZone(c.Zone),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	if err := storage.CreateNewBucket(exocmd.GContext, c.Bucket, c.ACL); err != nil {
		return fmt.Errorf("unable to create bucket: %w", err)
	}

	if !globalstate.Quiet {
		return utils.PrintOutput(storage.ShowBucket(exocmd.GContext, c.Bucket))
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storageMbCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
