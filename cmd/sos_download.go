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
		bucket := args[0]
		objectName := args[1]
		localFilePath := args[2]

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

		// Verify if destination already exists.
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			exists, err := destinationExists(localFilePath, objectName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("file %q: already exists", localFilePath)
			}
		}

		// Gather md5sum.
		objectStat, err := sosClient.StatObjectWithContext(gContext, bucket, objectName, minio.StatObjectOptions{})
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
		st, err := filePart.Stat()
		if err != nil {
			return err
		}

		object, err := sosClient.GetObjectWithContext(gContext, bucket, objectName, minio.GetObjectOptions{})
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

		progress := mpb.NewWithContext(gContext,
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

// destinationExists verifies that the given destination does not exist
func destinationExists(localFilePath string, objectName string) (bool, error) {
	st, err := os.Stat(localFilePath)
	if err == nil {
		// If the destination exists and is a directory.
		if st.IsDir() {
			localFilePath = path.Join(localFilePath, objectName)
			_, err = os.Stat(localFilePath)
			if err == nil {
				return true, nil
			}
		} else {
			return true, nil
		}
	}

	// Proceed if file does not exist. return for all other errors.
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
	}
	return false, nil
}

func init() {
	sosCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().BoolP("force", "f", false, "Overwrite the destination if it already exists")
}
