package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var storageSetACLCmd = &cobra.Command{
	Use:   "setacl sos://BUCKET/[OBJECT|PREFIX/] [CANNED-ACL]",
	Short: "Set a bucket/objects ACL",
	Long: fmt.Sprintf(`This command sets bucket/objects ACL.
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
https://community.exoscale.com/documentation/storage/acl/

If you want to target objects under a "directory" prefix, suffix the path
argument with "/":

    exo storage setacl sos://my-bucket/ --full-control alice@example.net
    exo storage setacl -r sos://my-bucket/public/ public-read

Supported output template annotations:

	* When showing a bucket: %s
	* When showing an object: %s`,
		strings.Join(sos.BucketCannedACLToStrings(), ", "),
		strings.Join(sos.ObjectCannedACLToStrings(), ", "),
		strings.Join(output.OutputterTemplateAnnotations(&sos.ShowBucketOutput{}), ", "),
		strings.Join(output.OutputterTemplateAnnotations(&storageShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		if (len(args) == 2 && storageACLFromCmdFlags(cmd.Flags()) != nil) ||
			(len(args) == 1 && storageACLFromCmdFlags(cmd.Flags()) == nil) {
			cmdExitOnUsageError(cmd, "either a canned ACL or ACL grantee options must be specified")
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
			acl    *storageACL
		)

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			prefix = parts[1]

			// Special case: the caller wants to target objects at the root of
			// the bucket, in this case the prefix is empty so we set it to a
			// symbolic value that shall be removed later on.
			if prefix == "" {
				prefix = "/"
			}
		}

		storage, err := sos.NewStorageClient(
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if acl = storageACLFromCmdFlags(cmd.Flags()); acl == nil {
			acl = &storageACL{Canned: args[1]}
		}

		if prefix == "" {
			if err := storage.SetBucketACL(bucket, acl); err != nil {
				return fmt.Errorf("unable to set ACL: %w", err)
			}

			if !globalstate.Quiet {
				return printOutput(storage.ShowBucket(bucket))
			}
			return nil
		}

		if err := storage.SetObjectsACL(bucket, prefix, acl, recursive); err != nil {
			return fmt.Errorf("unable to set ACL: %w", err)
		}

		if !globalstate.Quiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return printOutput(storage.ShowObject(bucket, prefix))
		}

		if !globalstate.Quiet {
			fmt.Println("ACL set successfully")
		}
		return nil
	},
}

func init() {
	storageSetACLCmd.Flags().BoolP("recursive", "r", false,
		"set ACL recursively (with object prefix only)")
	storageSetACLCmd.Flags().String(storageSetACLCmdFlagRead, "", "ACL Read grantee")
	storageSetACLCmd.Flags().String(storageSetACLCmdFlagWrite, "", "ACP Write grantee")
	storageSetACLCmd.Flags().String(storageSetACLCmdFlagReadACP, "", "ACP Read ACP grantee")
	storageSetACLCmd.Flags().String(storageSetACLCmdFlagWriteACP, "", "ACP Write ACP grantee")
	storageSetACLCmd.Flags().String(storageSetACLCmdFlagFullControl, "", "ACP Full Control grantee")
	storageCmd.AddCommand(storageSetACLCmd)
}

// storageACLFromCmdFlags returns a non-nil pointer to a storageACL struct if at least
// one of the ACL-related command flags is set.
func storageACLFromCmdFlags(flags *pflag.FlagSet) *storageACL {
	var acl *storageACL

	flags.VisitAll(func(flag *pflag.Flag) {
		switch flag.Name {
		case storageSetACLCmdFlagRead:
			if v := flag.Value.String(); v != "" {
				if acl == nil {
					acl = &storageACL{}
				}

				acl.Read = v
			}

		case storageSetACLCmdFlagWrite:
			if v := flag.Value.String(); v != "" {
				if acl == nil {
					acl = &storageACL{}
				}

				acl.Write = v
			}

		case storageSetACLCmdFlagReadACP:
			if v := flag.Value.String(); v != "" {
				if acl == nil {
					acl = &storageACL{}
				}

				acl.ReadACP = v
			}

		case storageSetACLCmdFlagWriteACP:
			if v := flag.Value.String(); v != "" {
				if acl == nil {
					acl = &storageACL{}
				}

				acl.WriteACP = v
			}

		case storageSetACLCmdFlagFullControl:
			if v := flag.Value.String(); v != "" {
				if acl == nil {
					acl = &storageACL{}
				}

				acl.FullControl = v
			}

		default:
			return
		}
	})

	return acl
}
