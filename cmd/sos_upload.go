package cmd

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"

	minio "github.com/minio/minio-go"
	"github.com/spf13/cobra"
)

type fileToUpload struct {
	localPath   string
	remotePath  string
	contentType string
}

// uploadCmd represents the upload command
var sosUploadCmd = &cobra.Command{
	Use:     "upload <bucket name> <local file path>+",
	Short:   "Upload an object into a bucket",
	Aliases: gUploadAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		remoteFilePath, err := cmd.Flags().GetString("remote-path")
		if err != nil {
			return err
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

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		// Upload the  file
		filesToUpload := []fileToUpload{}
		bucketName := args[0]

		for _, arg := range args[1:] {

			arg = filepath.ToSlash(arg)
			objectName := filepath.Base(arg)
			filePath := arg

			remote := strings.TrimLeft(filepath.ToSlash(remoteFilePath), "/")

			if strings.HasSuffix(remote, "/") {
				remote = remoteFilePath + objectName
			}

			if remote == "" {
				remote = objectName
			}

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}

			fileStat, err := file.Stat()
			if err != nil {
				return err
			}

			if recursive && fileStat.IsDir() {
				filesToUpload, err = getFiles(filePath, strings.TrimRight(remote, "/"), filesToUpload)
			} else {
				// Only the first 512 bytes are used to sniff the content type.
				buffer := make([]byte, 512)
				_, err = file.Read(buffer)

				contentType := http.DetectContentType(buffer)
				filesToUpload = append(filesToUpload, fileToUpload{
					localPath:   filePath,
					remotePath:  remote,
					contentType: contentType,
				})
			}
			if err != nil {
				return err
			}

			if err = file.Close(); err != nil {
				return err
			}

		}

		for _, fileToUpload := range filesToUpload {

			fileInfo, err := os.Stat(fileToUpload.localPath)
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

			base := filepath.Base(fileToUpload.localPath)

			bar := progress.AddBar(fileInfo.Size(),
				mpb.AppendDecorators(
					// simple name decorator
					decor.Name(base, decor.WC{W: len(base) + 1, C: decor.DidentRight}),
				),
				mpb.PrependDecorators(
					decor.AverageETA(decor.ET_STYLE_GO),
					// decor.DSyncWidth bit enables column width synchronization
					decor.Percentage(decor.WCSyncSpace),
				),
			)

			f, err := os.Open(fileToUpload.localPath)
			if err != nil {
				return err
			}

			reader := bar.ProxyReader(f)

			// Upload object with FPutObject
			_, err = minioClient.PutObjectWithContext(gContext, bucketName, fileToUpload.remotePath, f, fileInfo.Size(), minio.PutObjectOptions{ContentType: fileToUpload.contentType, Progress: reader})
			if err != nil {
				return err
			}

			progress.Wait()

			if err := f.Close(); err != nil {
				return err
			}

		}

		return nil
	},
}

func getFiles(folderName, remoteFilePath string, resFiles []fileToUpload) ([]fileToUpload, error) {
	files, err := ioutil.ReadDir(folderName)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		localPath := filepath.Join(folderName, f.Name())
		if f.IsDir() {
			resFiles, err = getFiles(localPath, filepath.Join(remoteFilePath, f.Name()), resFiles)
			if err != nil {
				return nil, err
			}
			continue
		}

		file, err := os.Open(localPath)
		if err != nil {
			return nil, err
		}

		var contentType string
		if f.Size() >= 512 {
			// Only the first 512 bytes are used to sniff the content type.
			buffer := make([]byte, 512)
			_, err = file.Read(buffer)
			if err != nil {
				return nil, err
			}

			contentType = http.DetectContentType(buffer)
		}

		resFiles = append(resFiles, fileToUpload{
			localPath:   localPath,
			remotePath:  filepath.Join(remoteFilePath, f.Name()),
			contentType: contentType,
		})

		if err := file.Close(); err != nil {
			return nil, err
		}
	}
	return resFiles, nil
}

func init() {
	sosCmd.AddCommand(sosUploadCmd)
	sosUploadCmd.Flags().BoolP("recursive", "r", false, "Upload a folder recursively")
	sosUploadCmd.Flags().StringP("remote-path", "p", "", "Set a remote path for local file(s)")
}
