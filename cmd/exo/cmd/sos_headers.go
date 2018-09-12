package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale/cmd/exo/table"
	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

// nolint
const (
	contentType = iota
	cacheControl
	contentEncoding
	contentDisposition
	contentLanguage
	expires
)

var supportedHeaders = []string{
	"content-type",
	"cache-control",
	"content-encoding",
	"content-disposition",
	"content-language",
	"expires",
}

// headersCmd represents the headers command
var sosHeadersCmd = &cobra.Command{
	Use:   "header",
	Short: "Object headers management",
}

func init() {
	sosCmd.AddCommand(sosHeadersCmd)
}

// headersCmd represents the headers command
var sosAddHeadersCmd = &cobra.Command{
	Use:   "add <bucket name> <object name>",
	Short: "Add an header key/value to an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		meta, err := getHeaderFlags(cmd)
		if err != nil {
			return err
		}

		if len(meta) == 0 {
			println("error: You have to choose one flag")
			if err = cmd.Usage(); err != nil {
				return err
			}
			return fmt.Errorf("error: You have to choose one flag")
		}

		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			return err
		}

		location, err := minioClient.GetBucketLocation(args[0])
		if err != nil {
			return err
		}

		minioClient, err = newMinioClient(location)
		if err != nil {
			return err
		}

		objInfo, err := minioClient.GetObjectACL(args[0], args[1])
		if err != nil {
			return err
		}

		_, ok := meta["content-type"]
		if !ok {
			objInfo.Metadata.Add("content-type", objInfo.ContentType)
		}

		src := minio.NewSourceInfo(args[0], args[1], nil)

		mergeHeader(src.Headers, objInfo.Metadata)

		// Destination object
		dst, err := minio.NewDestinationInfo(args[0], args[1], nil, meta)
		if err != nil {
			return err
		}

		// Copy object call
		return minioClient.CopyObject(dst, src)
	},
}

func getHeaderFlags(cmd *cobra.Command) (map[string]string, error) {
	res := map[string]string{}

	for i := 0; i <= expires; i++ {
		key, err := cmd.Flags().GetString(supportedHeaders[i])
		if err != nil {
			return nil, err
		}
		if key != "" {
			res[supportedHeaders[i]] = key
		}
	}
	return res, nil
}

func init() {
	sosHeadersCmd.AddCommand(sosAddHeadersCmd)
	sosAddHeadersCmd.Flags().SortFlags = false
	for i := 0; i <= expires; i++ {
		sosAddHeadersCmd.Flags().StringP(
			supportedHeaders[i],
			"",
			"",
			fmt.Sprintf("Add %s with <key>", strings.Replace(supportedHeaders[i], "-", " ", -1)))
	}

}

// headersCmd represents the headers command
var sosRemoveHeadersCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove an header key/value from an object",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		meta, err := getHeaderBool(cmd)
		if err != nil {
			return err
		}

		if meta == nil {
			println("error: You have to choose one flag")
			if err = cmd.Usage(); err != nil {
				return err
			}
			return fmt.Errorf("error: You have to choose one flag")
		}

		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			return err
		}

		location, err := minioClient.GetBucketLocation(args[0])
		if err != nil {
			return err
		}

		minioClient, err = newMinioClient(location)
		if err != nil {
			return err
		}

		objInfo, err := minioClient.GetObjectACL(args[0], args[1])
		if err != nil {
			return err
		}

		objInfo.Metadata.Add("content-type", objInfo.ContentType)
		objInfo.Metadata.Add("x-amz-metadata-directive", "REPLACE")

		for _, v := range meta {
			objInfo.Metadata.Del(v)
		}

		src := minio.NewSourceInfo(args[0], args[1], nil)

		mergeHeader(src.Headers, objInfo.Metadata)

		// Destination object
		dst, err := minio.NewDestinationInfo(args[0], args[1], nil, nil)
		if err != nil {
			return err
		}

		// Copy object call
		return minioClient.CopyObject(dst, src)
	},
}

func getHeaderBool(cmd *cobra.Command) ([]string, error) {
	var res []string

	for i := 1; i <= expires; i++ {
		key, err := cmd.Flags().GetBool(supportedHeaders[i])
		if err != nil {
			return nil, err
		}
		if key {
			res = append(res, supportedHeaders[i])
		}
	}
	return res, nil
}

func init() {
	sosHeadersCmd.AddCommand(sosRemoveHeadersCmd)
	sosRemoveHeadersCmd.Flags().SortFlags = false
	for i := 1; i <= expires; i++ {
		sosRemoveHeadersCmd.Flags().BoolP(
			supportedHeaders[i],
			"",
			false,
			fmt.Sprintf("Remove %s with <key>", strings.Replace(supportedHeaders[i], "-", " ", -1)))
	}
}

// headersCmd represents the headers command
var sosShowHeadersCmd = &cobra.Command{
	Use:     "list <bucket name> <object name>",
	Short:   "list object headers",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			return err
		}

		location, err := minioClient.GetBucketLocation(args[0])
		if err != nil {
			return err
		}

		minioClient, err = newMinioClient(location)
		if err != nil {
			return err
		}

		objInfo, err := minioClient.GetObjectACL(args[0], args[1])
		if err != nil {
			return err
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"File Name", "Key", "Value"})

		if objInfo.ContentType != "" {
			table.Append([]string{objInfo.Key, "content-type", objInfo.ContentType})
		}

		for k, v := range objInfo.Metadata {
			k = strings.ToLower(k)
			if isStandardHeader(k) && len(v) > 0 {
				table.Append([]string{objInfo.Key, k, v[0]})
			}
		}

		table.Render()

		return nil
	},
}

func isStandardHeader(headerKey string) bool {
	key := strings.ToLower(headerKey)
	for _, header := range supportedHeaders {
		if strings.ToLower(header) == key {
			return true
		}
	}
	return false
}

func init() {
	sosHeadersCmd.AddCommand(sosShowHeadersCmd)
}
