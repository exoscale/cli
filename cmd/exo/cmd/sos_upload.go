package cmd

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"

	humanize "github.com/dustin/go-humanize"
	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

// uploadCmd represents the upload command
var sosUploadCmd = &cobra.Command{
	Use:     "upload <bucket name> <local file path> [remote file path]",
	Short:   "Upload an object into a bucket",
	Aliases: gUploadAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		args[1] = filepath.ToSlash(args[1])

		var remoteFilePath string
		if len(args) > 2 {
			remoteFilePath = strings.TrimLeft(filepath.ToSlash(args[2]), "/")
		}

		minioClient, err := newMinioClient(sosZone)
		if err != nil {
			return err
		}

		location, errGetBucket := minioClient.GetBucketLocation(args[0])
		if errGetBucket != nil {
			return err
		}

		minioClient, err = newMinioClient(location)
		if err != nil {
			return err
		}

		// Upload the  file
		bucketName := args[0]
		objectName := filepath.Base(args[1])
		filePath := args[1]

		if strings.HasSuffix(remoteFilePath, "/") {
			remoteFilePath = remoteFilePath + objectName
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}

		// Only the first 512 bytes are used to sniff the content type.
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil {
			return err
		}

		if err = file.Close(); err != nil {
			return err
		}

		if remoteFilePath == "" {
			remoteFilePath = objectName
		}

		contentType := http.DetectContentType(buffer)

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		progress := mpb.New(
			mpb.WithContext(gContext),
			// override default (80) width
			mpb.WithWidth(64),
			// override default 120ms refresh rate
			mpb.WithRefreshRate(180*time.Millisecond),
		)

		bar := progress.AddBar(fileInfo.Size(),
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

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close() // nolint: errcheck

		reader := bar.ProxyReader(f)

		// Upload object with FPutObject
		n, err := minioClient.PutObjectWithContext(gContext, bucketName, remoteFilePath, f, fileInfo.Size(), minio.PutObjectOptions{ContentType: contentType, Progress: reader})
		if err != nil {
			return err
		}

		progress.Wait()

		log.Printf("Successfully uploaded %s of size %s\n", objectName, humanize.IBytes(uint64(n)))

		return nil
	},
}

func init() {
	sosCmd.AddCommand(sosUploadCmd)
}
