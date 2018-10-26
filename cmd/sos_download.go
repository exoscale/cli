package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download <bucket name> <object name> <file path>",
	Short: "Download an object from a bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return cmd.Usage()
		}

		bucketName := args[0]
		objectName := args[1]
		localFilePath := args[2]

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

		// Verify if destination already exists.
		st, err := os.Stat(localFilePath)
		if err == nil {
			// If the destination exists and is a directory.
			if st.IsDir() {
				return fmt.Errorf("file path is a directory")
			}
		}

		// Proceed if file does not exist. return for all other errors.
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}

		// Gather md5sum.
		objectStat, err := minioClient.StatObject(bucketName, objectName, minio.StatObjectOptions{})
		if err != nil {
			return err
		}

		// Write to a temporary file "fileName.part.minio" before saving.
		filePartPath := localFilePath + objectStat.ETag + ".part.minio"

		// If exists, open in append mode. If not create it as a part file.
		filePart, err := os.OpenFile(filePartPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

		// Issue Stat to get the current offset.
		st, err = filePart.Stat()
		if err != nil {
			return err
		}

		opts := minio.GetObjectOptions{}
		// Initialize get object request headers to set the
		// appropriate range offsets to read from.
		if st.Size() > 0 {
			opts.SetRange(st.Size(), 0)
		}

		object, err := minioClient.GetObjectWithContext(gContext, bucketName, objectName, opts)
		if err != nil {
			return err
		}
		defer object.Close() //nolint: errcheck

		progress := mpb.New(
			mpb.WithContext(gContext),
			// override default (80) width
			mpb.WithWidth(64),
			// override default 120ms refresh rate
			mpb.WithRefreshRate(180*time.Millisecond),
		)

		bar := progress.AddBar(objectStat.Size,
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(objectName, decor.WC{W: len(objectName) + 1, C: decor.DidentRight}),
				// decor.DSyncWidth bit enables column width synchronization
				decor.Percentage(decor.WCSyncSpace),
			),
			mpb.AppendDecorators(
				decor.AverageETA(decor.ET_STYLE_GO),
			),
		)

		reader := bar.ProxyReader(object)

		// Write to the part file.
		if _, err = io.CopyN(filePart, reader, objectStat.Size); err != nil {
			return err
		}

		// Close the file before rename, this is specifically needed for Windows users.
		if err = filePart.Close(); err != nil {
			return err
		}

		// Safely completed. Now commit by renaming to actual filename.
		if err = os.Rename(filePartPath, localFilePath); err != nil {
			return err
		}

		progress.Wait()

		log.Printf("Successfully downloaded %s into %q\n", args[1], args[2])
		return nil
	},
}

func init() {
	sosCmd.AddCommand(downloadCmd)
}
