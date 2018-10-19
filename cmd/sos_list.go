package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	humanize "github.com/dustin/go-humanize"
	minio "github.com/minio/minio-go"

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
		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			log.Fatal(err)
		}

		isRec, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		isRaw, err := cmd.Flags().GetBool("raw")
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return displayBucket(minioClient, isRec, isRaw)
		}

		path := filepath.ToSlash(args[0])
		path = strings.Trim(path, "/")
		p := splitPath(args[0])

		if len(p) == 0 || p[0] == "" {
			return displayBucket(minioClient, isRec, isRaw)
		}

		var prefix string
		if len(p) > 1 {
			prefix = path[len(p[0]):]
			prefix = strings.Trim(prefix, "/")
		}

		bucketName := p[0]

		///XXX waiting for pithos 301 redirect
		location, err := minioClient.GetBucketLocation(bucketName)
		if err != nil {
			return err
		}

		minioClient, err = newMinioClient(location)
		if err != nil {
			return err
		}
		///

		table := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

		if isRec {
			listRecursively(minioClient, bucketName, prefix, "", false, isRaw, table)
			return table.Flush()
		}

		recursive := true

		last := ""
		for message := range minioClient.ListObjectsV2(bucketName, prefix, recursive, gContext.Done()) {
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

			if isRaw {
				fmt.Println(key)
				continue
			}

			fmt.Fprintf(table, "%s\t%s\t%s\n", // nolint: errcheck
				fmt.Sprintf("[%s]", lastModified),
				fmt.Sprintf("%6s ", humanize.IBytes(uint64(message.Size))),
				key) // nolint: errcheck
		}

		return table.Flush()
	},
}

func listRecursively(c *minio.Client, bucketName, prefix, zone string, displayBucket, isRaw bool, table io.Writer) {

	for message := range c.ListObjectsV2(bucketName, prefix, true, gContext.Done()) {
		sPrefix := splitPath(prefix)
		sKey := splitPath(message.Key)
		if len(prefix) != len(strings.Join(sKey[:len(sPrefix)], "/")) {
			continue
		}

		lastModified := message.LastModified.Format(printDate)
		if displayBucket {
			if isRaw {
				fmt.Printf("%s/%s\n", bucketName, message.Key)
				continue
			}
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", fmt.Sprintf("[%s]", lastModified), // nolint: errcheck
				fmt.Sprintf("[%s]", zone),
				fmt.Sprintf("%6s ", humanize.IBytes(uint64(message.Size))),
				fmt.Sprintf("%s/%s", bucketName, message.Key)) // nolint: errcheck
		} else {
			if isRaw {
				fmt.Printf("%s\n", message.Key)
				continue
			}
			fmt.Fprintf(table, "%s\t%s\t%s\n", fmt.Sprintf("[%s]", lastModified), // nolint: errcheck
				fmt.Sprintf("%6s ", humanize.IBytes(uint64(message.Size))),
				message.Key) // nolint: errcheck
		}
	}
}

func displayBucket(minioClient *minio.Client, isRecursive, isRaw bool) error {
	allBuckets, err := listBucket(minioClient)
	if err != nil {
		return err
	}

	table := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	for zoneName, buckets := range allBuckets {
		for _, bucket := range buckets {
			if isRecursive {
				///XXX Waiting for pithos 301 redirect
				minioClient, err = newMinioClient(zoneName)
				if err != nil {
					return err
				}
				///
				if isRaw {
					fmt.Printf("%s/\n", bucket.Name)
				} else {
					fmt.Fprintf(table, "%s\t%s\t%s\t%s/\n", // nolint: errcheck
						fmt.Sprintf("[%s]", bucket.CreationDate.Format(printDate)),
						fmt.Sprintf("[%s]", zoneName),
						fmt.Sprintf("%6s ", humanize.IBytes(uint64(0))),
						bucket.Name) // nolint: errcheck
				}
				listRecursively(minioClient, bucket.Name, "", zoneName, true, isRaw, table)
				continue
			}
			if isRaw {
				fmt.Printf("%s/\n", bucket.Name)
				continue
			}
			fmt.Fprintf(table, "%s\t%s\t%s\t%s/\n", // nolint: errcheck
				fmt.Sprintf("[%s]", bucket.CreationDate.Format(printDate)),
				fmt.Sprintf("[%s]", zoneName),
				fmt.Sprintf("%6s ", humanize.IBytes(uint64(0))),
				bucket.Name) // nolint: errcheck
		}
	}
	return table.Flush()
}

func listBucket(minioClient *minio.Client) (map[string][]minio.BucketInfo, error) {
	bucketInfos, err := minioClient.ListBuckets()
	if err != nil {
		return nil, err
	}

	res := map[string][]minio.BucketInfo{}

	for _, bucketInfo := range bucketInfos {

		bucketLocation, err := minioClient.GetBucketLocation(bucketInfo.Name)
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
	sosListCmd.Flags().BoolP("raw", "l", false, "List without long format")
}
