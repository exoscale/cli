package cmd

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"

	"github.com/exoscale/cli/pkg/globalstate"

	minio "github.com/minio/minio-go/v6"
	"github.com/spf13/cobra"
)

const (
	parallelSosUpload = 10
)

type fileToUpload struct {
	localPath   string
	remotePath  string
	contentType string
}

var sosUploadCmd = &cobra.Command{
	Use:     "upload BUCKET FILE...",
	Short:   "Upload a file into a bucket",
	Aliases: gUploadAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		files := args[1:]

		remoteFilePath, err := cmd.Flags().GetString("remote-path")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		logfile, err := cmd.Flags().GetString("log")
		if err != nil {
			return err
		}
		if logfile != "" {
			l, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			sosClient.TraceOn(l)
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		// Upload the  file
		filesToUpload := []fileToUpload{}

		for _, file := range files {
			file = filepath.ToSlash(file)
			objectName := filepath.Base(file)
			filePath := file

			remote := strings.TrimLeft(filepath.ToSlash(remoteFilePath), "/")

			if (len(files) > 1 && remote != "") || (len(files) == 1 && strings.HasSuffix(remote, "/")) {
				remote = filepath.Join(remoteFilePath, objectName)
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
				var contentType string
				var bufferSize int64

				// Only the first 512 bytes are used to sniff the content type.
				if fileStat.Size() >= 512 {
					bufferSize = 512
				} else {
					bufferSize = fileStat.Size()
				}

				buffer := make([]byte, bufferSize)
				_, err = file.Read(buffer)

				contentType = http.DetectContentType(buffer)

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

		lenFileToUpload := len(filesToUpload)

		var taskWG sync.WaitGroup
		progress := mpb.NewWithContext(gContext,
			mpb.WithWaitGroup(&taskWG),
			// override default (80) width
			mpb.WithWidth(64),
			// override default 120ms refresh rate
			mpb.WithRefreshRate(180*time.Millisecond),
			mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool { return globalstate.Quiet }),
		)
		taskWG.Add(lenFileToUpload)

		workerSem := make(chan int, parallelSosUpload)

		for _, fileToUP := range filesToUpload {
			fileToUP := fileToUP
			go func() {
				defer taskWG.Done()
				workerSem <- 1

				fileInfo, err := os.Stat(fileToUP.localPath)
				if err != nil {
					log.Fatal(err)
				}

				f, err := os.Open(fileToUP.localPath)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close() //nolint: errcheck

				base := filepath.Base(fileToUP.localPath)
				bar := progress.AddBar(fileInfo.Size(),
					mpb.AppendDecorators(
						// simple name decorator
						decor.Name(base, decor.WC{W: len(base) + 1, C: decor.DidentRight}),
					),
					mpb.PrependDecorators(
						decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO), "done!"),
						// decor.DSyncWidth bit enables column width synchronization
						decor.Percentage(decor.WCSyncSpace),
					),
				)

				reader := bar.ProxyReader(f)
				// Upload object with FPutObject
				_, upErr := sosClient.PutObjectWithContext(
					gContext,
					bucket,
					fileToUP.remotePath,
					reader,
					fileInfo.Size(),
					minio.PutObjectOptions{
						ContentType:    fileToUP.contentType,
						SendContentMd5: true,
					},
				)
				if upErr != nil {
					log.Fatal(upErr)
				}

				// Workaround required to avoid the io.Reader from hanging when uploading empty files
				// (see https://github.com/vbauerster/mpb/issues/7#issuecomment-518756758)
				if fileInfo.Size() == 0 {
					bar.SetTotal(100, true)
				}

				<-workerSem
			}()
		}

		progress.Wait()
		return nil
	},
}

func getFiles(folderName, remoteFilePath string, resFiles []fileToUpload) ([]fileToUpload, error) {
	files, err := os.ReadDir(folderName)
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

		finfo, err := f.Info()
		if err != nil {
			return nil, err
		}

		var contentType string
		if finfo.Size() >= 512 {
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
	sosUploadCmd.Flags().StringP("log", "l", "", "Log upload transfer details to file")
	sosUploadCmd.Flags().BoolP("recursive", "r", false, "Upload a folder recursively")
	sosUploadCmd.Flags().StringP("remote-path", "p", "", "Set a remote path for local file(s)")
}
