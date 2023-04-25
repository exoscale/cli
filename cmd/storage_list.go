package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

var storageListCmd = &cobra.Command{
	Use:   "list [sos://BUCKET[/[PREFIX/]]",
	Short: "List buckets and objects",
	Long: fmt.Sprintf(`This command lists buckets and their objects.

If no argument is passed, this commands lists existing buckets. If a prefix is
specified (e.g. "sos://my-bucket/.../") the command lists the objects stored
in the bucket under the corresponding prefix.

Supported output template annotations:

  * When listing buckets: %s
  * When listing objects: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&sos.ListBucketsItemOutput{}), ", "),
		strings.Join(output.OutputterTemplateAnnotations(&sos.ListObjectsItemOutput{}), ", ")),
	Aliases: gListAlias,

	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
		}
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		if len(args) == 0 {
			return printOutput(listStorageBuckets())
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		stream, err := cmd.Flags().GetBool("stream")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			prefix = parts[1]
		}

		storage, err := sos.NewStorageClient(
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		return printOutput(storage.ListObjects(bucket, prefix, recursive, stream))
	},
}

func init() {
	storageListCmd.Flags().BoolP("recursive", "r", false,
		"list bucket recursively")
	storageListCmd.Flags().BoolP("stream", "s", false,
		"stream listed files instead of waiting for complete listing (useful for large buckets)")
	storageCmd.AddCommand(storageListCmd)
}

func listStorageBuckets() (output.Outputter, error) {
	out := make(storageListBucketsOutput, 0)

	res, err := cs.RequestWithContext(gContext, egoscale.ListBucketsUsage{})
	if err != nil {
		return nil, err
	}

	for _, b := range res.(*egoscale.ListBucketsUsageResponse).BucketsUsage {
		created, err := time.Parse(time.RFC3339, b.Created)
		if err != nil {
			return nil, err
		}

		out = append(out, sos.ListBucketsItemOutput{
			Name:    b.Name,
			Zone:    b.Region,
			Size:    b.Usage,
			Created: created.Format(sos.TimestampFormat),
		})
	}

	return &out, nil
}
