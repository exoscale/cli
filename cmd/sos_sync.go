package cmd

import (
	"errors"
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

//region Settings
const (
	defaultParallelSosSync = 10
)

//endregion

//region Type definitions
type sosSyncClientFactory = func(certsFile string) (*sosClient, error)
type sosSyncListObjects = func(config sosSyncConfiguration, errors chan<- error) <-chan sosSyncObject
type sosSyncListFiles = func(config sosSyncConfiguration, errors chan<- error) <-chan sosSyncFile
type sosSyncGetFile = func(config sosSyncConfiguration, file string) (sosSyncFile, error)
type sosSyncDiff = func(sosSyncConfiguration sosSyncConfiguration, done chan bool, errors chan<- error) <-chan sosSyncTask
type sosSyncProcessTaskList = func(sosSyncConfiguration sosSyncConfiguration, tasks <-chan sosSyncTask, done chan bool, errors chan<- error)
type sosSyncFileUi struct {
	getReader func(r io.Reader) io.ReadCloser
	error     func()
	complete  func()
}
type sosSyncFileUiFactory = func(filename string, filesize int64) sosSyncFileUi
type sosSyncUi = func(wg *sync.WaitGroup) sosSyncFileUiFactory

type sosSyncConfiguration struct {
	RemoveDeleted   bool
	DryRun          bool
	SourceDirectory string
	TargetBucket    string
	TargetPath      string
	Concurrency     uint16
}
type sosSyncObject struct {
	Key          string
	LastModified time.Time
	ContentType  string
	Size         int64
}
type sosSyncFile struct {
	Path         string
	LastModified time.Time
	Size         int64
}

const (
	sosSyncUploadAction = 0
	sosSyncDeleteAction = 1
)

type sosSyncTask struct {
	Action int
	File   string
	Size   int64
}

//endregion

//region Implementation
func newSosSyncCobraCommand(sosClientFactory sosSyncClientFactory) *cobra.Command {
	runE := newSosSyncRunE(sosClientFactory)
	return &cobra.Command{
		Use:     "sync <bucket name> <local path> <remote-path>",
		Short:   "Sync a local folder with the object storage",
		Aliases: gSyncAlias,
		RunE:    runE,
	}
}

func newSosSyncRunE(sosClientFactory sosSyncClientFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		removeDeleted, err := cmd.Flags().GetBool("remove-deleted")
		if err != nil {
			return err
		}
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		concurrency, err := cmd.Flags().GetUint16("concurrency")
		if err != nil {
			return err
		}
		if concurrency < 1 {
			return errors.New("concurrency cannot be less than 1")
		}

		if len(args) < 3 {
			return cmd.Usage()
		}
		targetBucket := args[0]

		sourceDirectory, err := filepath.Abs(args[1])
		if err != nil {
			return err
		}

		targetPath := strings.Trim(args[2], "/") + "/"

		sosClient, err := sosClientFactory(certsFile)
		if err != nil {
			return err
		}

		if _, err := os.Stat(sourceDirectory); os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("Source dirctory does not exist: %s", sourceDirectory))
		}

		var bucketExists bool
		bucketExists, err = sosClient.BucketExists(targetBucket)
		if err != nil {
			return err
		}
		if !bucketExists {
			return errors.New(fmt.Sprintf("The target bucket does not exist: %s", targetBucket))
		}

		location, err := sosClient.GetBucketLocation(targetBucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		var config = sosSyncConfiguration{
			RemoveDeleted:   removeDeleted,
			DryRun:          dryRun,
			TargetBucket:    targetBucket,
			SourceDirectory: filepath.ToSlash(sourceDirectory),
			TargetPath:      strings.Trim(targetPath, "/"),
			Concurrency:     concurrency,
		}

		err = sosSyncProcess(
			config,
			newSosSyncDiff(
				newSosSyncListObjects(sosClient),
				newSosSyncGetFile(),
				newSosSyncListFiles(),
			),
			newSosSyncProcessTaskList(
				sosClient,
				newSosSyncUi(),
			),
		)
		if err != nil {
			return err
		}

		return nil
	}
}

func newSosSyncUi() sosSyncUi {
	return func(wg *sync.WaitGroup) sosSyncFileUiFactory {
		progress := mpb.NewWithContext(gContext,
			mpb.WithWaitGroup(wg),
			// override default (80) width
			mpb.WithWidth(64),
			// override default 120ms refresh rate
			mpb.WithRefreshRate(180*time.Millisecond),
			mpb.ContainerOptOnCond(mpb.WithOutput(nil), func() bool { return gQuiet }),
		)
		return func(filename string, filesize int64) sosSyncFileUi {
			bar := progress.AddBar(filesize,
				mpb.AppendDecorators(
					// simple name decorator
					decor.Name(filename, decor.WC{W: len(filename) + 1, C: decor.DidentRight}),
				),
				mpb.PrependDecorators(
					decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO), "done!"),
					// decor.DSyncWidth bit enables column width synchronization
					decor.Percentage(decor.WCSyncSpace),
				),
			)
			return sosSyncFileUi{
				getReader: func(r io.Reader) io.ReadCloser {
					return bar.ProxyReader(r)
				},
				complete: func() {
					bar.Completed()
					if filesize == 0 {
						bar.SetTotal(100, true)
					}
				},
				error: func() {
					bar.Abort(false)
				},
			}
		}
	}
}

func newSosSyncListObjects(sosClient *sosClient) sosSyncListObjects {
	return func(config sosSyncConfiguration, errorChannel chan<- error) <-chan sosSyncObject {
		result := make(chan sosSyncObject)

		go func() {
			doneCh := make(chan struct{})
			defer close(doneCh)
			//todo how does ListObjectsV2 indicate errors?
			objects := sosClient.ListObjectsV2(config.TargetBucket, config.TargetPath, true, doneCh)
			for object := range objects {
				result <- sosSyncObject{
					Key:          object.Key[len(config.TargetPath):],
					Size:         object.Size,
					LastModified: object.LastModified,
					ContentType:  object.ContentType,
				}
			}
			close(result)
		}()

		return result
	}
}

func newSosSyncGetFile() sosSyncGetFile {
	return func(config sosSyncConfiguration, file string) (sosSyncFile, error) {
		trimmedFile := strings.Trim(file, "/")
		stat, err := os.Stat(config.SourceDirectory + "/" + trimmedFile)
		if err != nil {
			return sosSyncFile{}, err
		}

		return sosSyncFile{
			Path:         trimmedFile,
			Size:         stat.Size(),
			LastModified: stat.ModTime(),
		}, nil
	}
}

func newSosSyncListFiles() sosSyncListFiles {
	return func(config sosSyncConfiguration, errorChannel chan<- error) <-chan sosSyncFile {
		result := make(chan sosSyncFile)

		go func() {
			defer close(result)
			walkErr := filepath.Walk(config.SourceDirectory,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						result <- sosSyncFile{
							Path:         filepath.ToSlash(path[len(config.SourceDirectory)+1:]),
							LastModified: info.ModTime(),
							Size:         info.Size(),
						}
					}
					return nil
				})
			if walkErr != nil {
				errorChannel <- walkErr
			}
		}()
		return result
	}
}

func newSosSyncDiff(
	sosSyncListObjects sosSyncListObjects,
	sosSyncGetFile sosSyncGetFile,
	sosSyncListFiles sosSyncListFiles,
) sosSyncDiff {
	return func(sosSyncConfiguration sosSyncConfiguration, done chan bool, errorChannel chan<- error) <-chan sosSyncTask {
		result := make(chan sosSyncTask)

		go func() {
			defer close(result)

			remoteObjectsIndexed := map[string]sosSyncObject{}
			remoteObjects := sosSyncListObjects(sosSyncConfiguration, errorChannel)
			for remoteObject := range remoteObjects {
				_, err := sosSyncGetFile(sosSyncConfiguration, remoteObject.Key)
				if err != nil {
					if sosSyncConfiguration.RemoveDeleted {
						result <- sosSyncTask{
							File:   remoteObject.Key,
							Action: sosSyncDeleteAction,
						}
					}
				} else {
					remoteObjectsIndexed[remoteObject.Key] = remoteObject
				}
			}

			localFiles := sosSyncListFiles(sosSyncConfiguration, errorChannel)
			for localFile := range localFiles {
				if _, ok := remoteObjectsIndexed[localFile.Path]; ok != true {
					result <- sosSyncTask{
						File:   localFile.Path,
						Action: sosSyncUploadAction,
						Size:   0,
					}
				} else {
					if remoteObjectsIndexed[localFile.Path].LastModified.Before(localFile.LastModified) ||
						remoteObjectsIndexed[localFile.Path].Size != localFile.Size {
						result <- sosSyncTask{
							File:   localFile.Path,
							Action: sosSyncUploadAction,
							Size:   localFile.Size,
						}
					}
				}
			}

			if done != nil {
				done <- true
			}
		}()
		return result
	}
}

//todo this could probably do with a refactor
func newSosSyncProcessTaskList(sosClient *sosClient, ui sosSyncUi) sosSyncProcessTaskList {
	return func(sosSyncConfiguration sosSyncConfiguration, tasks <-chan sosSyncTask, inputDone chan bool, errorChannel chan<- error) {
		var taskWG sync.WaitGroup
		workerSem := make(chan int, sosSyncConfiguration.Concurrency)
		uploadUi := ui(&taskWG)
		for task := range tasks {
			taskWG.Add(1)
			task := task
			go func() {
				defer taskWG.Done()
				workerSem <- 1
				taskUi := uploadUi(task.File, task.Size)
				remotePath := strings.Trim(sosSyncConfiguration.TargetPath+"/"+task.File, "/")
				if task.Action == sosSyncDeleteAction {
					if sosSyncConfiguration.DryRun {
						fmt.Printf("[Dry run] Pretending to delete remote file: %s\n", remotePath)
					} else {
						fmt.Printf("Deleting remote file: %s\n", remotePath)
					}
					if !sosSyncConfiguration.DryRun {
						err := sosClient.RemoveObject(sosSyncConfiguration.TargetBucket, remotePath)
						if err != nil {
							taskUi.error()
							errorChannel <- err
						} else {
							taskUi.complete()
						}
					} else {
						taskUi.complete()
					}
				} else {
					localPath := filepath.FromSlash(sosSyncConfiguration.SourceDirectory + "/" + task.File)
					if sosSyncConfiguration.DryRun {
						fmt.Printf("[Dry run] Pretending to upload local file %s to remote path %s\n", localPath, remotePath)
					} else {
						fmt.Printf("Uploading local file %s to remote path %s\n", localPath, remotePath)
					}

					//region MIME type
					var contentType string
					var bufferSize int64

					// Only the first 512 bytes are used to sniff the content type.
					if task.Size >= 512 {
						bufferSize = 512
					} else {
						bufferSize = task.Size
					}

					file, err := os.Open(localPath)
					if err != nil {
						errorChannel <- err
					} else {
						buffer := make([]byte, bufferSize)
						_, err = file.Read(buffer)
						if err != nil {
							taskUi.error()
							errorChannel <- err
						} else {
							contentType = http.DetectContentType(buffer)
							//endregion

							//region Upload
							if !sosSyncConfiguration.DryRun {
								f, err := os.Open(localPath)
								if err != nil {
									taskUi.error()
									errorChannel <- err
								} else {
									//noinspection GoUnhandledErrorResult
									defer f.Close()

									_, upErr := sosClient.PutObjectWithContext(
										gContext,
										sosSyncConfiguration.TargetBucket,
										remotePath,
										taskUi.getReader(f),
										task.Size,
										minio.PutObjectOptions{
											ContentType: contentType,
										},
									)
									if upErr != nil {
										taskUi.error()
										errorChannel <- upErr
									} else {
										taskUi.complete()
									}
								}
							} else {
								taskUi.complete()
							}
							//endregion
						}
					}
				}
				<-workerSem
			}()
		}

		//Wait for complete input before waiting for the task WG
		<-inputDone
		close(inputDone)
		taskWG.Wait()
	}
}

//Takes a preconfigured object storage client for the target bucket, and a validated configuration
//and synchronizes the local folder with the remote.
func sosSyncProcess(
	sosSyncConfiguration sosSyncConfiguration,
	sosSyncDiff sosSyncDiff,
	sosSyncProcessTaskList sosSyncProcessTaskList,
) error {
	done := make(chan bool, 1)
	errorChannel := make(chan error)
	taskList := sosSyncDiff(sosSyncConfiguration, done, errorChannel)

	sosSyncProcessTaskList(sosSyncConfiguration, taskList, done, errorChannel)
	var err error = nil
	for {
		select {
		case errChannelResult := <-errorChannel:
			//noinspection GoUnhandledErrorResult
			fmt.Fprintf(os.Stderr, "Error while processing: %s", errChannelResult)
			err = errChannelResult
		default:
			err = nil
		}
		if err == nil {
			break
		}
	}
	if //noinspection GoNilness
	err != nil {
		return errors.New("one or more errors happened while processing your sync")
	} else {
		return nil
	}
}

//endregion

//region Wiring
func sosSyncLiveClientFactory(certsFile string) (*sosClient, error) {
	client, err := newSOSClient(certsFile)
	if err != nil {
		return nil, err
	}
	return client, err
}

func init() {
	cmd := newSosSyncCobraCommand(sosSyncLiveClientFactory)
	cmd.Flags().BoolP("remove-deleted", "r", false, "Remove remote files not present locally")
	cmd.Flags().BoolP("dry-run", "n", false, "Don't actually modify files")
	cmd.Flags().Uint16P("concurrency", "c", defaultParallelSosSync, "Parallel threads to use for upload")
	sosCmd.AddCommand(cmd)
}

//endregion
