package cmd

import (
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	minio "github.com/minio/minio-go/v6"
	"github.com/spf13/cobra"
)

// metadataCmd represents the metadata command
var sosMetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Object metadata management",
}

func init() {
	sosCmd.AddCommand(sosMetadataCmd)
}

// metadataCmd represents the metadata command
var sosAddMetadataCmd = &cobra.Command{
	Use:   "add <bucket name> <object name> <key> <value>",
	Short: "Add metadata to an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 4 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]
		k, v := args[2], args[3]

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
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

		objInfo, err := sosClient.GetObjectACL(bucket, object)
		if err != nil {
			return err
		}

		objInfo.Metadata.Add("content-type", objInfo.ContentType)

		src := minio.NewSourceInfo(bucket, object, nil)

		mergeHeader(src.Headers, objInfo.Metadata)

		meta := map[string]string{k: v}

		// Destination object
		dst, err := minio.NewDestinationInfo(bucket, object, nil, meta)
		if err != nil {
			return err
		}

		// Copy object call
		return sosClient.CopyObject(dst, src)
	},
}

func init() {
	sosMetadataCmd.AddCommand(sosAddMetadataCmd)
}

// metadataCmd represents the metadata command
var sosRemoveMetadataCmd = &cobra.Command{
	Use:     "remove <bucket name> <object name> <key>",
	Aliases: gRemoveAlias,
	Short:   "Remove metadata from an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
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

		objInfo, err := sosClient.GetObjectACL(bucket, object)
		if err != nil {
			return err
		}

		objInfo.Metadata.Add("content-type", objInfo.ContentType)
		objInfo.Metadata.Add("x-amz-metadata-directive", "REPLACE")

		for k := range objInfo.Metadata {
			key := strings.ToLower(k)
			if strings.HasPrefix(key, "x-amz-meta-") && strings.HasSuffix(key, k) {
				objInfo.Metadata.Del(k)
			}
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

func init() {
	sosMetadataCmd.AddCommand(sosRemoveMetadataCmd)
}

// metadataCmd represents the metadata command
var sosShowMetadataCmd = &cobra.Command{
	Use:     "list <bucket name> <object name>",
	Short:   "List object metadata",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
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

		objInfo, err := sosClient.GetObjectACL(bucket, object)
		if err != nil {
			return err
		}

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"File Name", "Key", "Value"})

		for k, v := range objInfo.Metadata {
			k = strings.ToLower(k)
			if strings.HasPrefix(k, "x-amz-meta-") && len(v) > 0 {
				table.Append([]string{objInfo.Key, k[len("x-amz-meta-"):], v[0]})
			}
		}

		table.Render()

		return nil
	},
}

func init() {
	sosMetadataCmd.AddCommand(sosShowMetadataCmd)
}
