package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

const (
	printDate = "2006-01-02 15:04:05 MST"
)

// sosListCmd represents the list command
var sosListCmd = &cobra.Command{
	Use:   "list [<bucket name>/path]",
	Short: "List file and folder",
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

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return displayBucket(sosClient, isRecursive, isShort)
		}

		path := filepath.ToSlash(args[0])
		path = strings.Trim(path, "/")
		p := splitPath(args[0])

		if len(p) == 0 || p[0] == "" {
			return displayBucket(sosClient, isRecursive, isShort)
		}

		var prefix string
		if len(p) > 1 {
			prefix = path[len(p[0]):]
			prefix = strings.Trim(prefix, "/")
		}

		bucket := p[0]

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		table := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

		if isRecursive {
			listRecursively(sosClient, bucket, prefix, "", false, isShort, table)
			return table.Flush()
		}

		recursive := true

		last := ""
		for message := range sosClient.ListObjectsV2(bucket, prefix, recursive, gContext.Done()) {
			if message.Err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", message.Err)
				continue
			}

			sPrefix := splitPath(prefix)
			sKey := splitPath(message.Key)

			// dont display if prefix == object path or folder
			if isPrefix(prefix, message.Key) {
				continue
			}

			// skip if message key contain prefix and is not equal
			if len(prefix) != len(strings.Join(sKey[:len(sPrefix)], "/")) {
				continue
			}

			// detect all file in current work dir (keep folder using last var)
			if len(sKey) > len(sPrefix)+1 && sKey[len(sPrefix)] == last {
				last = sKey[len(sPrefix)]
				continue
			}
			last = sKey[len(sPrefix)]

			// if is a folder (format key)
			if len(sKey) > len(sPrefix)+1 {
				message.Key = strings.Join([]string{prefix, sKey[len(sPrefix)]}, "/") + "/"
				message.Size = 0
			}

			lastModified := message.LastModified.Format(printDate)
			key := filepath.ToSlash(message.Key)
			key = strings.TrimLeft(key[len(prefix):], "/")

			if isShort {
				fmt.Fprintln(table, key) // nolint: errcheck
				continue
			}

			fmt.Fprintf(table, "[%s]\t%6s \t%s\n", lastModified, humanize.IBytes(uint64(message.Size)), key) // nolint: errcheck
		}

		return table.Flush()
	},
}

func listRecursively(sosClient *sosClient, bucket, prefix, zone string,
	displayBucket, isShort bool, table io.Writer) {
	for message := range sosClient.ListObjectsV2(bucket, prefix, true, gContext.Done()) {
		if message.Err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", message.Err)
			continue
		}

		sPrefix := splitPath(prefix)
		sKey := splitPath(message.Key)
		if len(prefix) != len(strings.Join(sKey[:len(sPrefix)], "/")) {
			continue
		}

		lastModified := message.LastModified.Format(printDate)
		var bucket string
		var zoneFormat string
		if displayBucket {
			bucket = fmt.Sprintf("%s/", bucket)
			zoneFormat = fmt.Sprintf("[%s]\t", zone)
		}
		if isShort {
			fmt.Fprintf(table, "%s%s\n", bucket, message.Key) // nolint: errcheck
			continue
		}
		fmt.Fprintf(table,
			"[%s]\t%s%6s \t%s%s\n", lastModified, zoneFormat, humanize.IBytes(uint64(message.Size)), bucket, message.Key) // nolint: errcheck
	}
}

func displayBucket(sosClient *sosClient, isRecursive, isShort bool) error {
	resp, err := cs.RequestWithContext(gContext, egoscale.ListBucketsUsage{})
	if err != nil {
		return err
	}

	buckets := resp.(*egoscale.ListBucketsUsageResponse)

	table := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	for _, b := range buckets.BucketsUsage {
		if isShort {
			fmt.Fprintf(table, "%s/\n", b.Name) // nolint: errcheck
		} else {
			t, err := time.Parse(time.RFC3339, b.Created)
			if err != nil {
				return err
			}

			fmt.Fprintf(table,
				"[%s]\t[%s]\t%6s \t%s/\n", t.Format(printDate), b.Region, humanize.IBytes(uint64(b.Usage)), b.Name) // nolint: errcheck
		}
		if isRecursive {
			if err = sosClient.setZone(b.Region); err != nil {
				return err
			}
			listRecursively(sosClient, b.Name, "", b.Region, true, isShort, table)
		}
	}
	return table.Flush()
}

func isPrefix(prefix, file string) bool {
	prefix = strings.Trim(prefix, "/")
	file = strings.Trim(file, "/")
	return prefix == file
}

func splitPath(s string) []string {
	path := filepath.ToSlash(s)
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func init() {
	sosCmd.AddCommand(sosListCmd)
	sosListCmd.Flags().BoolP("recursive", "r", false, "List recursively")
	sosListCmd.Flags().BoolP("short", "S", false, "List in short format")
}
