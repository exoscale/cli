package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

type storageUploadConfig struct {
	bucket    string
	prefix    string
	acl       string
	recursive bool
	dryRun    bool
}

var storageUploadCmd = &cobra.Command{
	Use:     "upload FILE... sos://BUCKET/[PREFIX/]",
	Aliases: []string{"put"},
	Short:   "Upload files to a bucket",
	Long: `This command uploads local files to a bucket.

Examples:

    # Upload files at the root of the bucket
    exo storage upload a b c sos://my-bucket

    # Upload files in a directory (trailing "/" in destination)
    exo storage upload index.html sos://my-bucket/public/

    # Upload a file under a different name
    exo storage upload a.txt sos://my-bucket/z.txt

    # Upload a directory recursively
    exo storage upload -r my-files/ sos://my-bucket
`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[len(args)-1] = strings.TrimPrefix(args[len(args)-1], storageBucketPrefix)

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string

			sources = args[:len(args)-1]
			dst     = args[len(args)-1]
		)

		acl, err := cmd.Flags().GetString("acl")
		if err != nil {
			return err
		}
		if acl != "" && !isInList(s3ObjectCannedACLToStrings(), acl) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl, strings.Join(s3ObjectCannedACLToStrings(), ", "))
		}

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		dstParts := strings.SplitN(dst, "/", 2)
		bucket = dstParts[0]
		if len(dstParts) > 1 {
			// Tricky case: if the user specifies "<bucket>/" as destination,
			// strings.SplitN()'s result slice contains an empty string as last
			// item: in this case we set the prefix as "/" to mean the root of
			// the bucket.
			if dstParts[len(dstParts)-1] == "" {
				prefix = "/"
			} else {
				prefix = dstParts[1]
			}
		} else {
			prefix = "/"
		}

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		return storage.uploadFiles(sources, &storageUploadConfig{
			bucket:    bucket,
			prefix:    prefix,
			acl:       acl,
			recursive: recursive,
			dryRun:    dryRun,
		})
	},
}

func init() {
	storageUploadCmd.Flags().String("acl", "",
		fmt.Sprintf("canned ACL to set on object (%s)", strings.Join(s3ObjectCannedACLToStrings(), "|")))
	storageUploadCmd.Flags().BoolP("dry-run", "n", false,
		"simulate files upload, don't actually do it")
	storageUploadCmd.Flags().BoolP("recursive", "r", false,
		"upload directories recursively")
	storageCmd.AddCommand(storageUploadCmd)
}

func (c *storageClient) uploadFiles(sources []string, config *storageUploadConfig) error {
	if len(sources) > 1 && !strings.HasSuffix(config.prefix, "/") {
		return errors.New(`multiple files to upload, destination must end with "/"`)
	}

	if config.dryRun {
		fmt.Println("[DRY-RUN]")
	}

	for _, src := range sources {
		src := src

		srcInfo, err := os.Stat(src)
		if err != nil {
			return err
		}

		if srcInfo.IsDir() {
			if !config.recursive {
				return fmt.Errorf("%q is a directory, use flag `-r` to upload recursively", src)
			}

			err = filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
				var (
					key    string
					prefix = config.prefix
				)

				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				/*
					Handle directory-type source similar to rsync. Considering the following source file tree:

					    my-dir/
					    ├── a
					    ├── b
					    └── x/
					       ├── y

					 Specifying "my-dir/" (with a trailing slash) will upload files such as:

					     bucket/a
					     bucket/b
					     bucket/x/y

					 Whereas specifying "my-dir" (without trailing slash) will upload files such as:

					     bucket/my-dir/a
					     bucket/my-dir/b
					     bucket/my-dir/x/y
				*/

				if strings.HasSuffix(src, "/") {
					key = strings.TrimPrefix(filePath, path.Clean(src)+"/")
				} else {
					key = path.Clean(filePath)
				}

				if prefix != "" {
					// The "/" value can be used at command-level to mean that we want to
					// list from the root of the bucket, but the actual bucket root is an
					// empty prefix.
					if prefix == "/" {
						prefix = ""
					}

					if prefix != "" {
						key = path.Join(prefix, key)
					}
				}

				if config.dryRun {
					fmt.Printf("%s -> %s/%s\n", src, config.bucket, key)
					return nil
				}

				return c.uploadFile(config.bucket, filePath, key, config.acl)
			})
			if err != nil {
				return err
			}
		} else {
			key := path.Base(src)

			if prefix := config.prefix; prefix != "" {
				// The "/" value can be used at command-level to mean that we want to
				// list from the root of the bucket, but the actual bucket root is an
				// empty prefix.
				if prefix == "/" {
					prefix = ""
				}

				if prefix != "" {
					if strings.HasSuffix(prefix, "/") {
						key = path.Join(prefix, key)
					} else {
						key = prefix
					}
				}
			}

			if config.dryRun {
				fmt.Printf("%s -> %s/%s\n", src, config.bucket, key)
				continue
			}

			if err := c.uploadFile(config.bucket, src, key, config.acl); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *storageClient) uploadFile(bucket, file, key, acl string) error {
	maxFilenameLen := 16

	pb := mpb.New(
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool {
			return gQuiet
		}),
	)

	file = path.Clean(file)

	fileInfo, err := os.Stat(file)
	if err != nil {
		return err
	}

	bar := pb.AddBar(fileInfo.Size(),
		mpb.PrependDecorators(
			decor.Name(ellipString(file, maxFilenameLen), decor.WC{W: maxFilenameLen, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f", decor.WCSyncWidthR),
			decor.Name(" | "),
			decor.Elapsed(decor.ET_STYLE_GO),
		),
	)

	f, err := os.Open(file)
	if err != nil {
		return err
	}

	putObjectInput := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bar.ProxyReader(f),
	}

	if acl != "" {
		putObjectInput.ACL = s3types.ObjectCannedACL(acl)
	}

	_, err = s3manager.
		NewUploader(c.Client).
		Upload(gContext, &putObjectInput)

	pb.Wait()

	return err
}
