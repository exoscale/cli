package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

type storageDownloadConfig struct {
	bucket      string
	prefix      string
	source      string
	destination string
	objects     []*s3types.Object
	recursive   bool
	overwrite   bool
	dryRun      bool
}

var storageDownloadCmd = &cobra.Command{
	Use:     "download sos://BUCKET/[OBJECT|PREFIX/] [DESTINATION]",
	Aliases: []string{"get"},
	Short:   "Download files from a bucket",
	Long: `This command downloads files from a bucket.

If no destination argument is provided, files will be stored into the current
directory.

Examples:

    # Download a single file
    exo storage download sos://my-bucket/file-a

    # Download a single file and rename it locally
    exo storage download sos://my-bucket/file-a file-z

    # Download a prefix recursively
    exo storage download -r sos://my-bucket/public/ /tmp/public/
`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		// Append implicit root prefix ("/") if only a bucket name is specified in the source
		if !strings.Contains(args[0], "/") {
			args[0] = args[0] + "/"
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string

			src = args[0]
			dst = "./"
		)

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(src, "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			prefix = parts[1]
			if prefix == "" {
				prefix = "/"
			}
		}

		if len(args) == 2 {
			dst = args[1]
		}

		if strings.HasSuffix(src, "/") && !recursive {
			return fmt.Errorf("%q is a directory, use flag `-r` to download recursively", src)
		}

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		objects := make([]*s3types.Object, 0)
		if err := storage.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
			objects = append(objects, o)
			return nil
		}); err != nil {
			return fmt.Errorf("error listing objects: %s", err)
		}

		return storage.downloadFiles(&storageDownloadConfig{
			bucket:      bucket,
			prefix:      prefix,
			source:      src,
			objects:     objects,
			destination: dst,
			recursive:   recursive,
			overwrite:   force,
			dryRun:      dryRun,
		})
	},
}

func init() {
	storageDownloadCmd.Flags().BoolP("force", "f", false,
		"overwrite existing destination files")
	storageDownloadCmd.Flags().BoolP("dry-run", "n", false,
		"simulate files download, don't actually do it")
	storageDownloadCmd.Flags().BoolP("recursive", "r", false,
		"download prefix recursively")
	storageCmd.AddCommand(storageDownloadCmd)
}

func (c *storageClient) downloadFiles(config *storageDownloadConfig) error {
	if len(config.objects) > 1 && !strings.HasSuffix(config.destination, "/") {
		return errors.New(`multiple files to download, destination must end with "/"`)
	}

	// Handle relative filesystem destination (e.g. ".", "../.." etc.)
	if dstInfo, err := os.Stat(config.destination); err == nil {
		if dstInfo.IsDir() && !strings.HasSuffix(config.destination, "/") {
			config.destination += "/"
		}
	}

	if config.dryRun {
		fmt.Println("[DRY-RUN]")
	}

	for _, object := range config.objects {
		dst := func() string {
			if strings.HasSuffix(config.source, "/") {
				return path.Join(config.destination, strings.TrimPrefix(aws.ToString(object.Key), config.prefix))
			}

			if strings.HasSuffix(config.destination, "/") {
				return path.Join(config.destination, path.Base(aws.ToString(object.Key)))
			}

			return path.Join(config.destination)
		}()

		if config.dryRun {
			fmt.Printf("%s/%s -> %s\n", config.bucket, aws.ToString(object.Key), dst)
			continue
		}

		if _, err := os.Stat(dst); err == nil && !config.overwrite {
			return fmt.Errorf("file %q already exists, use flag `-f` to overwrite", dst)
		}

		if _, err := os.Stat(path.Dir(dst)); errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(path.Dir(dst), 0o755); err != nil {
				return err
			}
		}

		if err := c.downloadFile(config.bucket, object, dst); err != nil {
			return err
		}
	}

	return nil
}

func (c *storageClient) downloadFile(bucket string, object *s3types.Object, dst string) error {
	maxFilenameLen := 16

	pb := mpb.NewWithContext(gContext,
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool {
			return gQuiet
		}),
	)

	bar := pb.AddBar(
		object.Size,
		mpb.PrependDecorators(
			decor.Name(ellipString(aws.ToString(object.Key), maxFilenameLen),
				decor.WC{W: maxFilenameLen, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f", decor.WCSyncWidthR),
			decor.Name(" | "),
			decor.Elapsed(decor.ET_STYLE_GO),
		),
	)

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	getObjectInput := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    object.Key,
	}

	_, err = s3manager.
		NewDownloader(c.Client).
		Download(
			gContext,
			// mpb doesn't natively support the io.WriteAt interface expected
			// by the s3manager.Download()'s w parameter, so we wrap in a shim
			// to be able to track the download progress. Trick inspired from
			// https://github.com/vbauerster/mpb/blob/v4/proxyreader.go
			&proxyWriterAt{
				wt:  f,
				bar: bar,
				iT:  time.Now(),
			},
			&getObjectInput,
		)

	pb.Wait()

	if errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "\rDownload interrupted by user\n")
		return nil
	}

	return err
}
