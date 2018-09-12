package cmd

import (
	"os"
	"strings"

	"github.com/exoscale/egoscale/cmd/exo/table"
	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

// metadataCmd represents the metadata command
var sosMetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Object metadate management",
}

func init() {
	sosCmd.AddCommand(sosMetadataCmd)
}

// metadataCmd represents the metadata command
var sosAddMetadataCmd = &cobra.Command{
	Use:   "add <bucket name> <object name> <key> <value>",
	Short: "Add a metadata to an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 4 {
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

		objInfo.Metadata.Add("content-type", objInfo.ContentType)

		src := minio.NewSourceInfo(args[0], args[1], nil)

		mergeHeader(src.Headers, objInfo.Metadata)

		meta := map[string]string{
			args[2]: args[3],
		}

		// Destination object
		dst, err := minio.NewDestinationInfo(args[0], args[1], nil, meta)
		if err != nil {
			return err
		}

		// Copy object call
		return minioClient.CopyObject(dst, src)
	},
}

func init() {
	sosMetadataCmd.AddCommand(sosAddMetadataCmd)
}

// metadataCmd represents the metadata command
var sosRemoveMetadataCmd = &cobra.Command{
	Use:     "remove <bucket name> <object name> <key>",
	Aliases: gRemoveAlias,
	Short:   "Remove a metadata from an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
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

		objInfo.Metadata.Add("content-type", objInfo.ContentType)
		objInfo.Metadata.Add("x-amz-metadata-directive", "REPLACE")

		for k := range objInfo.Metadata {
			key := strings.ToLower(k)
			if strings.HasPrefix(key, "x-amz-meta-") && strings.HasSuffix(key, args[2]) {
				objInfo.Metadata.Del(k)
			}
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

func init() {
	sosMetadataCmd.AddCommand(sosRemoveMetadataCmd)
}

// metadataCmd represents the metadata command
var sosShowMetadataCmd = &cobra.Command{
	Use:     "list <bucket name> <object name>",
	Short:   "list object metadatas",
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
