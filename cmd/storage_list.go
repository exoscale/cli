package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dustin/go-humanize"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type storageListObjectsItemOutput struct {
	Path         string `json:"name"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified,omitempty"`
	Dir          bool   `json:"dir"`
}

type storageListObjectsOutput []storageListObjectsItemOutput

func (o *storageListObjectsOutput) toJSON() { outputJSON(o) }
func (o *storageListObjectsOutput) toText() { outputText(o) }
func (o *storageListObjectsOutput) toTable() {
	table := tabwriter.NewWriter(os.Stdout,
		0,
		0,
		1,
		' ',
		tabwriter.TabIndent)
	defer table.Flush()

	for _, f := range *o {
		if f.Dir {
			_, _ = fmt.Fprintf(table, " \tDIR \t%s/\n", f.Path)
		} else {
			_, _ = fmt.Fprintf(table, "%s\t%6s \t%s\n", f.LastModified, humanize.IBytes(uint64(f.Size)), f.Path)
		}
	}
}

type storageListBucketsItemOutput struct {
	Name    string `json:"name"`
	Zone    string `json:"zone"`
	Size    int64  `json:"size"`
	Created string `json:"created"`
}

type storageListBucketsOutput []storageListBucketsItemOutput

func (o *storageListBucketsOutput) toJSON() { outputJSON(o) }
func (o *storageListBucketsOutput) toText() { outputText(o) }
func (o *storageListBucketsOutput) toTable() {
	table := tabwriter.NewWriter(os.Stdout,
		0,
		0,
		1,
		' ',
		tabwriter.TabIndent)
	defer table.Flush()

	for _, b := range *o {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%6s \t%s/\n",
			b.Created, b.Zone, humanize.IBytes(uint64(b.Size)), b.Name)
	}
}

var storageListCmd = &cobra.Command{
	Use:   "list [bucket[/[prefix/]]",
	Short: "List buckets and objects",
	Long: fmt.Sprintf(`This command lists buckets and their objects.

If no argument is passed, this commands lists existing buckets. If a prefix is
specified (e.g. "bucket/.../") the command lists the objects stored in the
bucket under the corresponding prefix.

Supported output template annotations:

  * When listing buckets: %s
  * When listing objects: %s`,
		strings.Join(outputterTemplateAnnotations(&storageListBucketsItemOutput{}), ", "),
		strings.Join(outputterTemplateAnnotations(&storageListObjectsItemOutput{}), ", ")),
	Aliases: gListAlias,

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		if len(args) == 0 {
			return output(listStorageBuckets())
		}

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
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

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		return output(storage.listObjects(bucket, prefix, recursive, stream))
	},
}

func init() {
	storageListCmd.Flags().BoolP("recursive", "r", false,
		"list bucket recursively")
	storageListCmd.Flags().BoolP("stream", "s", false,
		"stream listed files instead of waiting for complete listing (useful for large buckets)")
	storageCmd.AddCommand(storageListCmd)
}

func listStorageBuckets() (outputter, error) {
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

		out = append(out, storageListBucketsItemOutput{
			Name:    b.Name,
			Zone:    b.Region,
			Size:    b.Usage,
			Created: created.Format("2006-01-02 15:04:05 MST"),
		})
	}

	return &out, nil
}

func (c *storageClient) listObjects(bucket, prefix string, recursive, stream bool) (outputter, error) {
	dirs := make(map[string]struct{})
	out := make(storageListObjectsOutput, 0)

	var ct string
	for {
		res, err := c.ListObjectsV2(gContext, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: aws.String(ct),
		})
		if err != nil {
			return nil, err
		}
		ct = aws.ToString(res.NextContinuationToken)

		for _, o := range res.Contents {
			// If not invoked in recursive mode, split object keys on the "/" separator then return
			// a "directory" placeholder for the base prefix and hide objects "below" the prefix.
			parts := strings.SplitN(strings.TrimPrefix(aws.ToString(o.Key), prefix), "/", 2)
			if len(parts) > 1 && !recursive {
				dir := path.Base(parts[0])
				if _, ok := dirs[dir]; !ok {
					if stream {
						fmt.Println(dir + "/")
					} else {
						out = append(out, storageListObjectsItemOutput{
							Path: dir,
							Dir:  true,
						})
					}
					dirs[dir] = struct{}{}
				}
				continue
			}

			if stream {
				fmt.Println(aws.ToString(o.Key))
			} else {
				out = append(out, storageListObjectsItemOutput{
					Path:         aws.ToString(o.Key),
					Size:         o.Size,
					LastModified: o.LastModified.Format("2006-01-02 15:04:05 MST"),
				})
			}
		}

		if !res.IsTruncated {
			break
		}
	}

	return &out, nil
}
