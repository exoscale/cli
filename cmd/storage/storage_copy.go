package storage

import (
	"github.com/spf13/cobra"
)

var storageCopyCmd = &cobra.Command{
	Use:     "copy sos://BUCKET/[OBJECT|PREFIX/] sos://BUCKET/[OBJECT|PREFIX/]",
	Aliases: []string{"cp"},
	Short:   "Copy objects within a bucket or across buckets",
	Long: `Copy objects within a bucket or across buckets.

This command copies objects server-side without downloading them. Object
metadata, headers, and ACLs are preserved. Existing destination objects are
overwritten.

Multi-object prefix copies are processed serially. A trailing slash on the
source selects prefix mode; -r controls recursion into subdirectories.

Examples:

    exo storage copy sos://my-bucket/file-a sos://my-bucket/folder/file-a

    exo storage copy sos://my-bucket/file-a sos://other-bucket/file-a

    exo storage copy -r sos://my-bucket/prefix/ sos://other-bucket/prefix/

    exo storage copy -n sos://my-bucket/file-a sos://other-bucket/file-a
`,
	PreRunE: validateStorageTransferCommand,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStorageTransfer(cmd, args, false)
	},
}

func init() {
	storageCopyCmd.Flags().BoolP("dry-run", "n", false, "simulate the copy operation")
	storageCopyCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
	storageCopyCmd.Flags().BoolP("recursive", "r", false, "copy objects recursively")
	storageCopyCmd.Flags().BoolP("verbose", "v", false, "output copied objects")
	storageCopyCmd.Flags().Int("multipart-concurrency", 4, "number of concurrent parts for multipart copies")
	storageCmd.AddCommand(storageCopyCmd)
}
