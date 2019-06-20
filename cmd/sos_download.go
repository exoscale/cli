package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"time"

	minio "github.com/minio/minio-go/v6"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
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
				localFilePath = path.Join(localFilePath, objectName)
				_, err = os.Stat(localFilePath)
				if err == nil {
					return fmt.Errorf("file %q: already exists", localFilePath)
				}
			} else {
				return fmt.Errorf("file %q: already exists", localFilePath)
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

		object, err := minioClient.GetObjectWithContext(gContext, bucketName, objectName, minio.GetObjectOptions{})
		if err != nil {
			return err
		}
		defer object.Close() //nolint: errcheck

		// XXXXX I have to do that because this feature of minio-go doesn't working:
		// THAT:
		_, err = object.Seek(st.Size(), 0)
		if err != nil {
			return err
		}
		// INSTEAD OF:
		/*
			opts := minio.GetObjectOptions{}

			// Initialize get object request headers to set the
			// appropriate range offsets to read from.
			if st.Size() > 0 {
				opts.SetRange(st.Size(), 0)
			}

			object, err := minioClient.GetObjectWithContext(gContext, bucketName, objectName, opts)
		*/
		// XXXXXX

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

		bar.IncrBy(int(st.Size()))

		reader := bar.ProxyReader(object)

		// Write to the part file.
		if _, err = io.CopyN(filePart, reader, objectStat.Size-st.Size()); err != nil {
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

		log.Printf("Successfully downloaded %q into %q\n", objectName, localFilePath)
		return nil
	},
}

func init() {
	sosCmd.AddCommand(downloadCmd)
}
