package cmd

import (
	"fmt"
	"os"
	"strings"

	minio "github.com/minio/minio-go/v6"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/table"
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

var sosHeadersCmd = &cobra.Command{
	Use:   "header",
	Short: "Object headers management",
}

func init() {
	sosCmd.AddCommand(sosHeadersCmd)
}

var sosAddHeadersCmd = &cobra.Command{
	Use:   "add BUCKET OBJECT",
	Short: "Add an header key/value to an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

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

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		objInfo, err := sosClient.GetObjectACLWithContext(gContext, bucket, object)
		if err != nil {
			return err
		}

		_, ok := meta["content-type"]
		if !ok {
			objInfo.Metadata.Add("content-type", objInfo.ContentType)
		}

		src := minio.NewSourceInfo(bucket, object, nil)

		mergeHeader(src.Headers, objInfo.Metadata)

		// Destination object
		dst, err := minio.NewDestinationInfo(bucket, object, nil, meta)
		if err != nil {
			return err
		}

		// Copy object call
		return sosClient.CopyObject(dst, src)
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
			fmt.Sprintf("Add %s with KEY", strings.ReplaceAll(supportedHeaders[i], "-", " ")))
	}
}

var sosRemoveHeadersCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove an header key/value from an object",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

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

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		objInfo, err := sosClient.GetObjectACLWithContext(gContext, bucket, object)
		if err != nil {
			return err
		}

		objInfo.Metadata.Add("content-type", objInfo.ContentType)
		objInfo.Metadata.Add("x-amz-metadata-directive", "REPLACE")

		for _, v := range meta {
			objInfo.Metadata.Del(v)
		}

		src := minio.NewSourceInfo(bucket, object, nil)

		mergeHeader(src.Headers, objInfo.Metadata)

		// Destination object
		dst, err := minio.NewDestinationInfo(bucket, object, nil, nil)
		if err != nil {
			return err
		}

		// Copy object call
		return sosClient.CopyObject(dst, src)
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
			fmt.Sprintf("Remove %s with KEY", strings.ReplaceAll(supportedHeaders[i], "-", " ")))
	}
}

var sosShowHeadersCmd = &cobra.Command{
	Use:     "list BUCKET OBJECT",
	Short:   "list object headers",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		objInfo, err := sosClient.GetObjectACLWithContext(gContext, bucket, object)
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
