package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

const (
	objOwnershipOpArgIndex           = 0
	objOwnershipBucketArgIndex       = 1
	objOwnershipStatus               = "status"
	objOwnershipObjectWriter         = "object-writer"
	objOwnershipBucketOwnerEnforced  = "bucket-owner-enforced"
	objOwnershipBucketOwnerPreferred = "bucket-owner-preferred"
)

func init() {
	storageBucketObjectOwnershipCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketObjectOwnershipCmd)
}

var storageBucketObjectOwnershipCmd = &cobra.Command{
	// TODO
	Use:     "object-ownership {status,object-writer,bucket-owner-enforced,bucket-owner-preferred} sos://BUCKET",
	Aliases: []string{"oo"},
	Short:   "Manage the Object Ownership setting of a Storage Bucket",
	Long:    storageBucketObjectOwnershipCmdLongHelp(),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		permittedOps := make(map[string]struct{}, 4)
		permittedOps[objOwnershipStatus] = struct{}{}
		permittedOps[objOwnershipObjectWriter] = struct{}{}
		permittedOps[objOwnershipBucketOwnerEnforced] = struct{}{}
		permittedOps[objOwnershipBucketOwnerPreferred] = struct{}{}

		if _, ok := permittedOps[args[objOwnershipOpArgIndex]]; !ok {
			cmdExitOnUsageError(cmd, "invalid operation")
		}

		args[objOwnershipBucketArgIndex] = strings.TrimPrefix(args[objOwnershipBucketArgIndex], storageBucketPrefix)

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{zoneFlagLong})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		ownershipCommand := args[objOwnershipOpArgIndex]
		bucket := args[objOwnershipBucketArgIndex]

		fmt.Println(ownershipCommand)

		zone, err := cmd.Flags().GetString(zoneFlagLong)
		if err != nil {
			return err
		}

		storage, err := newStorageClient(
			storageClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		switch ownershipCommand {
		case objOwnershipStatus:
			return output(storage.GetBucketObjectOwnership(cmd.Context(), bucket))
		case objOwnershipObjectWriter:
			return storage.SetBucketObjectOwnership(cmd.Context(), bucket, ObjectOwnershipObjectWriter)
		case objOwnershipBucketOwnerPreferred:
			return storage.SetBucketObjectOwnership(cmd.Context(), bucket, ObjectOwnershipBucketOwnerPreferred)
		case objOwnershipBucketOwnerEnforced:
			return storage.SetBucketObjectOwnership(cmd.Context(), bucket, ObjectOwnershipBucketOwnerEnforced)
		}

		return fmt.Errorf("invalid operation")
	},
}

var storageBucketObjectOwnershipCmdLongHelp = func() string {
	// TODO
	return "Manage the Object Ownership setting of a Storage Bucket"
}

type storageBucketObjectOwnershipOutput struct {
	Bucket          string `json:"bucket"`
	ObjectOwnership string `json:"objectOwnership"`
}

func (o *storageBucketObjectOwnershipOutput) toJSON() { output.JSON(o) }
func (o *storageBucketObjectOwnershipOutput) toText() { output.Text(o) }
func (o *storageBucketObjectOwnershipOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Bucket Object Ownership"})

	t.Append([]string{"Bucket", o.Bucket})
	// TODO naming?
	t.Append([]string{"Object Ownership", o.ObjectOwnership})
}
