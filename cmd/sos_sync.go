package cmd

import (
	"errors"
	"fmt"
	"github.com/minio/minio-go/v6"
	"log"
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
	parallelSosSync = 10
)

//endregion

//region Type definitions
type sosSyncClientFactory = func(certsFile string) (*sosClient, error)
type sosSyncListObjects = func(config sosSyncConfiguration) <-chan sosSyncObject
type sosSyncListFiles = func(config sosSyncConfiguration) <-chan sosSyncFile
type sosSyncGetFile = func(config sosSyncConfiguration, file string) (sosSyncFile, error)
type sosSyncDiff = func(sosSyncConfiguration sosSyncConfiguration, done chan bool) <-chan sosSyncTask
type sosSyncProcessTaskList = func(sosSyncConfiguration sosSyncConfiguration, tasks <-chan sosSyncTask, done chan bool) error

type sosSyncConfiguration struct {
	RemoveDeleted   bool
	DryRun          bool
	SourceDirectory string
	TargetBucket    string
	TargetPath      string
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
		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		removeDeleted, err := cmd.Parent().Flags().GetBool("remove-deleted")
		if err != nil {
			return err
		}
		dryRun, err := cmd.Parent().Flags().GetBool("dry-run")
		if err != nil {
			return err
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
			),
		)
		if err != nil {
			return err
		}

		return nil
	}
}

func newSosSyncListObjects(sosClient *sosClient) sosSyncListObjects {
	return func(config sosSyncConfiguration) <-chan sosSyncObject {
		result := make(chan sosSyncObject)
		go func() {
			doneCh := make(chan struct{})
			defer close(doneCh)
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
	return func(config sosSyncConfiguration) <-chan sosSyncFile {
		result := make(chan sosSyncFile)

		//todo ignored error?
		go func() error {
			defer close(result)
			err := filepath.Walk(config.SourceDirectory,
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
			if err != nil {
				return err
			}
			return nil
		}()
		return result
	}
}

func newSosSyncDiff(
	sosSyncListObjects sosSyncListObjects,
	sosSyncGetFile sosSyncGetFile,
	sosSyncListFiles sosSyncListFiles,
) sosSyncDiff {
	return func(sosSyncConfiguration sosSyncConfiguration, done chan bool) <-chan sosSyncTask {
		result := make(chan sosSyncTask)

		go func() {
			remoteObjectsIndexed := map[string]sosSyncObject{}
			remoteObjects := sosSyncListObjects(sosSyncConfiguration)
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
			localFiles := sosSyncListFiles(sosSyncConfiguration)
			for localFile := range localFiles {
				if _, ok := remoteObjectsIndexed[localFile.Path]; ok != true {
					result <- sosSyncTask{
						File:   localFile.Path,
						Action: sosSyncUploadAction,
					}
				} else {
					if remoteObjectsIndexed[localFile.Path].LastModified.Before(localFile.LastModified) ||
						remoteObjectsIndexed[localFile.Path].Size != localFile.Size {
						result <- sosSyncTask{
							File:   localFile.Path,
							Action: sosSyncUploadAction,
						}
					}
				}
			}

			defer close(result)
			if done != nil {
				done <- true
			}
		}()
		return result
	}
}

//todo add UI
func newSosSyncProcessTaskList(sosClient *sosClient) sosSyncProcessTaskList {
	return func(sosSyncConfiguration sosSyncConfiguration, tasks <-chan sosSyncTask, done chan bool) error {
		var taskWG sync.WaitGroup
		workerSem := make(chan int, parallelSosSync)
		for task := range tasks {
			taskWG.Add(1)
			task := task
			go func() {
				defer taskWG.Done()
				workerSem <- 1
				remotePath := strings.Trim(sosSyncConfiguration.TargetPath+"/"+task.File, "/")
				if task.Action == sosSyncDeleteAction {
					fmt.Printf("Deleting remote file: %s\n", remotePath)
					err := sosClient.RemoveObject(sosSyncConfiguration.TargetBucket, remotePath)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					localPath := filepath.FromSlash(sosSyncConfiguration.SourceDirectory + "/" + task.File)
					fmt.Printf("Uploading local file %s to remote path %s\n", localPath, remotePath)
					fileInfo, err := os.Stat(localPath)
					if err != nil {
						log.Fatal(err)
					}

					//region MIME type
					var contentType string
					var bufferSize int64

					// Only the first 512 bytes are used to sniff the content type.
					if fileInfo.Size() >= 512 {
						bufferSize = 512
					} else {
						bufferSize = fileInfo.Size()
					}

					file, err := os.Open(localPath)
					if err != nil {
						log.Fatal(err)
					}
					buffer := make([]byte, bufferSize)
					_, err = file.Read(buffer)

					contentType = http.DetectContentType(buffer)
					//endregion

					f, err := os.Open(localPath)
					if err != nil {
						log.Fatal(err)
					}
					defer f.Close()

					_, upErr := sosClient.PutObjectWithContext(
						gContext,
						sosSyncConfiguration.TargetBucket,
						remotePath,
						f,
						fileInfo.Size(),
						minio.PutObjectOptions{
							ContentType: contentType,
						},
					)
					if upErr != nil {
						log.Fatal(upErr)
					}
				}
				<-workerSem
			}()
		}

		//Wait for complete input before waiting for the task WG
		<-done
		close(done)
		taskWG.Wait()
		return nil
	}
}

//Takes a preconfigured object storage client for the target bucket, and a validated configuration
//and synchronizes the local folder with the remote.
func sosSyncProcess(
	sosSyncConfiguration sosSyncConfiguration,
	sosSyncDiff sosSyncDiff,
	sosSyncProcessTaskList sosSyncProcessTaskList,
) error {
	var done chan bool
	if !sosSyncConfiguration.DryRun {
		done = make(chan bool, 1)
	} else {
		done = nil
	}
	taskList := sosSyncDiff(sosSyncConfiguration, done)

	if sosSyncConfiguration.DryRun {
		fmt.Println("Dry run enabled, only printing actions that would be taken.")
		for task := range taskList {
			if task.Action == sosSyncUploadAction {
				fmt.Println(fmt.Sprintf("Uploading %s...", task.File))
			} else {
				fmt.Println(fmt.Sprintf("Deleting %s...", task.File))
			}
		}
		return nil
	} else {
		return sosSyncProcessTaskList(sosSyncConfiguration, taskList, done)
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
	sosCmd.AddCommand(newSosSyncCobraCommand(sosSyncLiveClientFactory))
	sosCmd.Flags().StringP("log", "l", "", "Log sync transfer details to file")
	//todo flags defined here seem to be ignored
	sosCmd.Flags().BoolP("remove-deleted", "r", false, "Remove remote files not present locally")
	sosCmd.Flags().BoolP("dry-run", "n", false, "Don't actually modify files")
}

//endregion
