package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

type storageSetACLCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"setacl"`

	Target    string `cli-arg:"#" cli-usage:"sos://BUCKET/[OBJECT|PREFIX/]"`
	CannedACL string `cli-arg:"?" cli-usage:"CANNED-ACL"`

	Recursive   bool   `cli-short:"r" cli-usage:"set ACL recursively (with object prefix only)"`
	Read        string `cli-flag:"read" cli-usage:"ACP Read grantee"`
	Write       string `cli-flag:"write" cli-usage:"ACP Write grantee"`
	ReadACP     string `cli-flag:"read-acp" cli-usage:"ACP Read ACP grantee"`
	WriteACP    string `cli-flag:"write-acp" cli-usage:"ACP Write ACP grantee"`
	FullControl string `cli-flag:"full-control" cli-usage:"ACP Full Control grantee"`
}

func (c *storageSetACLCmd) CmdAliases() []string { return nil }
func (c *storageSetACLCmd) CmdShort() string     { return "Set a bucket/objects ACL" }
func (c *storageSetACLCmd) CmdLong() string {
	return fmt.Sprintf(`This command sets bucket/objects ACL.
It can be used in 2 (mutually exclusive) forms:

    * With a "canned" ACL:

        exo storage setacl sos://my-bucket public-read

    * With explicit Access Control Policy (ACP) grantees:

        exo storage setacl sos://my-bucket \
            --full-control alice@example.net \
            --read bob@example.net

Supported canned ACLs:

    * For buckets: %s
    * For objects: %s

In ACP mode, it is possible to use the following special values to reference
pre-defined groups:

    * ALL_USERS (as in "public-read" canned ACL)
    * AUTHENTICATED_USERS (as in "authenticated-read" canned ACL)

For more information on ACL, please refer to the Exoscale Storage
documentation:
https://community.exoscale.com/product/storage/object-storage/how-to/acl/

If you want to target objects under a "directory" prefix, suffix the path
argument with "/":

    exo storage setacl sos://my-bucket/ --full-control alice@example.net
    exo storage setacl -r sos://my-bucket/public/ public-read

Supported output template annotations:

	* When showing a bucket: %s
	* When showing an object: %s`,
		strings.Join(sos.BucketCannedACLToStrings(), ", "),
		strings.Join(sos.ObjectCannedACLToStrings(), ", "),
		strings.Join(output.TemplateAnnotations(&sos.ShowBucketOutput{}), ", "),
		strings.Join(output.TemplateAnnotations(&sos.ShowObjectOutput{}), ", "))
}

func (c *storageSetACLCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if _, _, err := parsePath(c.Target); err != nil {
		clicmd.CmdExitOnUsageError(cmd, err.Error())
	}

	// A canned ACL or a grantee flag only counts when its value is not empty.
	hasCannedACL := c.CannedACL != ""
	hasGranteeFlagsSet := c.Read != "" || c.Write != "" || c.ReadACP != "" || c.WriteACP != "" || c.FullControl != ""
	if hasCannedACL == hasGranteeFlagsSet {
		clicmd.CmdExitOnUsageError(cmd, "either a canned ACL or ACL grantee options must be specified")
	}

	return nil
}

func (c *storageSetACLCmd) CmdRun(_ *cobra.Command, _ []string) error {
	bucket, prefix, err := parsePath(c.Target)
	if err != nil {
		return err
	}

	// Special case: "sos://BUCKET/" targets the objects at the root of the
	// bucket, while "sos://BUCKET" targets the bucket itself. The prefix is
	// empty in both cases, so we set it to a symbolic value that shall be
	// removed later on.
	if prefix == "" && strings.HasSuffix(c.Target, "/") {
		prefix = "/"
	}

	storage, err := sos.NewStorageClient(
		exocmd.GContext,
		sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to initialize storage client: %w", err)
	}

	// CmdPreRun made sure that either grantees or a canned ACL are set.
	acl := &sos.ACL{Canned: c.CannedACL}
	if c.Read != "" || c.Write != "" || c.ReadACP != "" || c.WriteACP != "" || c.FullControl != "" {
		acl = &sos.ACL{
			Read:        c.Read,
			Write:       c.Write,
			ReadACP:     c.ReadACP,
			WriteACP:    c.WriteACP,
			FullControl: c.FullControl,
		}
	}

	if prefix == "" {
		if err := storage.SetBucketACL(exocmd.GContext, bucket, acl); err != nil {
			return fmt.Errorf("unable to set ACL: %w", err)
		}

		if !globalstate.Quiet {
			return utils.PrintOutput(storage.ShowBucket(exocmd.GContext, bucket))
		}
		return nil
	}

	if err := storage.SetObjectsACL(exocmd.GContext, bucket, prefix, acl, c.Recursive); err != nil {
		return fmt.Errorf("unable to set ACL: %w", err)
	}

	if !globalstate.Quiet && !c.Recursive && !strings.HasSuffix(prefix, "/") {
		return utils.PrintOutput(storage.ShowObject(exocmd.GContext, bucket, prefix))
	}

	if !globalstate.Quiet {
		fmt.Println("ACL set successfully")
	}
	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &storageSetACLCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
