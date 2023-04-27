package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

const (
	printDate = "2006-01-02 15:04:05 MST"
)

var sosListCmd = &cobra.Command{
	Use:   "list [BUCKET/PATH]",
	Short: "List buckets and files",
	Long: `This command lists all your buckets or all the files stored in the specified bucket.

Note: the buckets size reported is computed daily, it may not be the actual size at the time of listing.`,
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		isRecursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		isShort, err := cmd.Flags().GetBool("short")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return displayBuckets(sosClient, isRecursive, isShort)
		}

		return displayBucket(sosClient, filepath.ToSlash(args[0]), isRecursive, isShort)
	},
}

func displayBuckets(sosClient *sosClient, isRecursive, isShort bool) error {
	resp, err := globalstate.EgoscaleClient.RequestWithContext(gContext, egoscale.ListBucketsUsage{})
	if err != nil {
		return err
	}

	buckets := resp.(*egoscale.ListBucketsUsageResponse)

	table := tabwriter.NewWriter(os.Stdout, 10, 0, 1, ' ', tabwriter.TabIndent)

	for _, b := range buckets.BucketsUsage {
		if isShort {
			fmt.Fprintf(table, "%s\n", b.Name) // nolint: errcheck
		} else {
			t, err := time.Parse(time.RFC3339, b.Created)
			if err != nil {
				return err
			}

			fmt.Fprintf(table,
				"[%s]\t[%s]\t%s\t%s\t\n", t.Format(printDate), b.Region, humanize.IBytes(uint64(b.Usage)), b.Name) // nolint: errcheck
		}

		table.Flush()

		if isRecursive {
			if err = sosClient.setZone(b.Region); err != nil {
				return err
			}

			listObjects(sosClient, b.Name, "", isRecursive, isShort)
		}
	}

	return nil
}

func displayBucket(sosClient *sosClient, path string, isRecursive, isShort bool) error {
	isDir := strings.HasSuffix(path, "/")

	path = strings.Trim(path, "/")
	splitPath := strings.Split(path, "/")

	bucket := splitPath[0]

	prefix := filepath.Join(splitPath[1:]...)
	if isDir && len(prefix) > 1 {
		prefix = prefix + "/"
	}

	zone, err := sosClient.GetBucketLocation(bucket)
	if err != nil {
		return err
	}

	if err := sosClient.setZone(zone); err != nil {
		return err
	}

	listObjects(sosClient, bucket, prefix, isRecursive, isShort)

	return nil
}

func listObjects(sosClient *sosClient, bucket, prefix string, isRecursive, isShort bool) {
	table := tabwriter.NewWriter(os.Stdout, 10, 0, 1, ' ', tabwriter.TabIndent)

	for object := range sosClient.ListObjectsV2(bucket, prefix, isRecursive, gContext.Done()) {
		table.Flush()

		if object.Err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", object.Err)
			continue
		}

		if isShort {
			fmt.Fprintf(table, "%s/%s\n", bucket, object.Key) // nolint: errcheck
			continue
		}

		if object.LastModified.IsZero() {
			fmt.Fprintf(table,
				"%s\t%s\t%s/%s\t\n",
				strings.Repeat(" ", 25),
				"DIR",
				bucket, object.Key) // nolint: errcheck
			continue
		}

		fmt.Fprintf(table,
			"[%s]\t%s\t%s/%s\t\n",
			object.LastModified.Format(printDate),
			humanize.IBytes(uint64(object.Size)),
			bucket, object.Key) // nolint: errcheck
	}

	table.Flush()
}

func init() {
	sosCmd.AddCommand(sosListCmd)
	sosListCmd.Flags().BoolP("recursive", "r", false, "List recursively")
	sosListCmd.Flags().BoolP("short", "S", false, "List in short format")
}
