package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	humanize "github.com/dustin/go-humanize"
	minio "github.com/minio/minio-go/v6"

	"github.com/spf13/cobra"
)

const (
	printDate = "2006-01-02 15:04:05 MST"
)

// sosListCmd represents the list command
var sosListCmd = &cobra.Command{
	Use:     "list [<bucket name>/path]",
	Short:   "List file and folder",
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
	allBuckets, err := listBucket(sosClient)
	if err != nil {
		return err
	}

	table := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)

	for zoneName, buckets := range allBuckets {
		for _, bucket := range buckets {
			if isShort {
				fmt.Fprintf(table, "%s/\n", bucket.Name) // nolint: errcheck
			} else {
				fmt.Fprintf(table,
					"[%s]\t[%s]\t%6s \t%s/\n", bucket.CreationDate.Format(printDate), zoneName, humanize.IBytes(uint64(0)), bucket.Name) // nolint: errcheck
			}
			if isRecursive {
				if err = sosClient.setZone(zoneName); err != nil {
					return err
				}
				listRecursively(sosClient, bucket.Name, "", zoneName, true, isShort, table)
			}
		}
	}
	return table.Flush()
}

func listBucket(sosClient *sosClient) (map[string][]minio.BucketInfo, error) {
	bucketInfos, err := sosClient.ListBuckets()
	if err != nil {
		return nil, err
	}

	res := map[string][]minio.BucketInfo{}

	for _, bucketInfo := range bucketInfos {
		bucketLocation, err := sosClient.GetBucketLocation(bucketInfo.Name)
		if err != nil {
			return nil, err
		}
		if _, ok := res[bucketLocation]; !ok {
			res[bucketLocation] = []minio.BucketInfo{bucketInfo}
			continue
		}

		res[bucketLocation] = append(res[bucketLocation], bucketInfo)

	}
	return res, nil
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
