package sos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dustin/go-humanize"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

func (c *Client) DeleteObjects(ctx context.Context, bucket, prefix string, recursive bool) ([]s3types.DeletedObject, error) {
	deleteList := make([]s3types.ObjectIdentifier, 0)
	err := c.ForEachObject(ctx, bucket, prefix, recursive, func(o *s3types.Object) error {
		deleteList = append(deleteList, s3types.ObjectIdentifier{Key: o.Key})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects to delete: %w", err)
	}

	// The S3 DeleteObjects API call is limited to 1000 keys per call, as a
	// precaution we're batching deletes.
	maxKeys := 1000
	deleted := make([]s3types.DeletedObject, 0)

	for i := 0; i < len(deleteList); i += maxKeys {
		j := i + maxKeys
		if j > len(deleteList) {
			j = len(deleteList)
		}

		res, err := c.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &bucket,
			Delete: &s3types.Delete{Objects: deleteList[i:j]},
		})
		if err != nil {
			return nil, err
		}

		deleted = append(deleted, res.Deleted...)
	}

	return deleted, nil
}

func (c *Client) GenPresignedURL(ctx context.Context, method, bucket, key string, expires time.Duration) (string, error) {
	var (
		psURL *v4.PresignedHTTPRequest
		err   error
	)

	psClient := s3.NewPresignClient(c.s3Client, func(o *s3.PresignOptions) {
		if expires > 0 {
			o.Expires = expires
		}
	})

	switch method {
	case "get":
		psURL, err = psClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	case "put":
		psURL, err = psClient.PresignPutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	default:
		err = fmt.Errorf("unsupported method %q", method)
	}

	if err != nil {
		return "", err
	}

	return psURL.URL, nil
}

type DownloadConfig struct {
	bucket      string
	prefix      string
	source      string
	destination string
	objects     []*s3types.Object
	recursive   bool
	overwrite   bool
	dryRun      bool
}

func (c *Client) DownloadFiles(ctx context.Context, config *DownloadConfig) error {
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

		if err := c.DownloadFile(ctx, config.bucket, object, dst); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DownloadFile(ctx context.Context, bucket string, object *s3types.Object, dst string) error {
	maxFilenameLen := 16

	pb := mpb.NewWithContext(ctx,
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool {
			return globalstate.Quiet
		}),
	)

	bar := pb.AddBar(
		object.Size,
		mpb.PrependDecorators(
			decor.Name(utils.EllipString(aws.ToString(object.Key), maxFilenameLen),
				decor.WC{W: maxFilenameLen, C: decor.DidentRight}),
		),
		mpb.AppendDecorators(
			decor.CountersKibiByte("% .2f / % .2f", decor.WCSyncWidthR),
			decor.Name(" | "),
			decor.Elapsed(decor.ET_STYLE_GO),
		),
	)

	// Workaround required to avoid the io.Reader from hanging when uploading empty files
	// (see https://github.com/vbauerster/mpb/issues/7#issuecomment-518756758)
	if object.Size == 0 {
		bar.SetTotal(100, true)
	}

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
		NewDownloader(c.s3Client).
		Download(
			ctx,
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

type ListObjectsOutput []ListObjectsItemOutput

func (o *ListObjectsOutput) toJSON() { output.JSON(o) }
func (o *ListObjectsOutput) toText() { output.Text(o) }
func (o *ListObjectsOutput) toTable() {
	table := tabwriter.NewWriter(os.Stdout,
		0,
		0,
		1,
		' ',
		tabwriter.TabIndent)
	defer table.Flush()

	for _, f := range *o {
		if f.Dir {
			_, _ = fmt.Fprintf(table, " \tDIR \t%s\n", f.Path)
		} else {
			_, _ = fmt.Fprintf(table, "%s\t%6s \t%s\n", f.LastModified, humanize.IBytes(uint64(f.Size)), f.Path)
		}
	}
}

type ListObjectsItemOutput struct {
	Path         string `json:"name"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified,omitempty"`
	Dir          bool   `json:"dir"`
}

type ListBucketsOutput []ListBucketsItemOutput

func (o *ListBucketsOutput) toJSON() { output.JSON(o) }
func (o *ListBucketsOutput) toText() { output.Text(o) }
func (o *ListBucketsOutput) toTable() {
	table := tabwriter.NewWriter(os.Stdout,
		0,
		0,
		1,
		' ',
		tabwriter.TabIndent)
	defer table.Flush()

	for _, b := range *o {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%6s \t%s/\n",
			b.Created, b.Zone, humanize.IBytes(uint64(b.Size)), b.Name)
	}
}

type ListBucketsItemOutput struct {
	Name    string `json:"name"`
	Zone    string `json:"zone"`
	Size    int64  `json:"size"`
	Created string `json:"created"`
}

func (c *Client) ListObjects(ctx context.Context, bucket, prefix string, recursive, stream bool) (output.Outputter, error) {
	out := make(ListObjectsOutput, 0)
	dirs := make(map[string]struct{})     // for deduplication of common prefixes (folders)
	dirsOut := make(ListObjectsOutput, 0) // to separate common prefixes (folders) from objects (files)

	var ct string
	for {
		req := s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: aws.String(ct),
		}
		if !recursive {
			req.Delimiter = aws.String("/")
		}

		res, err := c.s3Client.ListObjectsV2(ctx, &req)
		if err != nil {
			return nil, err
		}
		ct = aws.ToString(res.NextContinuationToken)

		if !recursive {
			for _, cp := range res.CommonPrefixes {
				dir := aws.ToString(cp.Prefix)
				if _, ok := dirs[dir]; !ok {
					if stream {
						fmt.Println(dir)
					} else {
						dirsOut = append(dirsOut, ListObjectsItemOutput{
							Path: dir,
							Dir:  true,
						})
					}
					dirs[dir] = struct{}{}
				}
			}
		}

		for _, o := range res.Contents {
			if stream {
				fmt.Println(aws.ToString(o.Key))
			} else {
				out = append(out, ListObjectsItemOutput{
					Path:         aws.ToString(o.Key),
					Size:         o.Size,
					LastModified: o.LastModified.Format(TimestampFormat),
				})
			}
		}

		if !res.IsTruncated {
			break
		}
	}

	// To be user friendly, we are going to push dir records to the top of the output list
	if !stream && !recursive {
		out = append(dirsOut, out...)
	}

	return &out, nil
}

type StorageUploadConfig struct {
	bucket    string
	prefix    string
	acl       string
	recursive bool
	dryRun    bool
}

func (c *Client) UploadFiles(ctx context.Context, sources []string, config *StorageUploadConfig) error {
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

				return c.UploadFile(ctx, config.bucket, filePath, key, config.acl)
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

			if err := c.UploadFile(ctx, config.bucket, src, key, config.acl); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) UploadFile(ctx context.Context, bucket, file, key, acl string) error {
	maxFilenameLen := 16

	pb := mpb.NewWithContext(ctx,
		mpb.ContainerOptOn(mpb.WithOutput(nil), func() bool {
			return globalstate.Quiet
		}),
	)

	file = path.Clean(file)

	fileInfo, err := os.Stat(file)
	if err != nil {
		return err
	}

	bar := pb.AddBar(fileInfo.Size(),
		mpb.PrependDecorators(
			decor.Name(utils.EllipString(file, maxFilenameLen), decor.WC{W: maxFilenameLen, C: decor.DidentRight}),
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

	var contentType string
	if fileInfo.Size() >= 512 {
		buf := make([]byte, 512) // http.DetectContentType() only looks at the first 512 bytes of the file.
		if _, err = f.Read(buf); err != nil {
			return err
		}
		contentType = http.DetectContentType(buf)
		if _, err = f.Seek(0, 0); err != nil {
			return err
		}
	}

	// Because we wrap the input with a ProxyReader to render a progress bar
	// The AWS SDK cannot perform PartSize estimation (we lose the io.Seeker implementation it relies on)
	// We therefore replicate that logic here, and explicitly set a part size to avoid
	// bumping into the s3manager.MaxUploadParts limit
	partSize, err := c.EstimatePartSize(f)
	if err != nil {
		return err
	}
	partSizeOpt := func(u *s3manager.Uploader) {
		u.PartSize = partSize
	}

	putObjectInput := s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bar.ProxyReader(f),
		ContentType: aws.String(contentType),
	}

	if acl != "" {
		putObjectInput.ACL = s3types.ObjectCannedACL(acl)
	}

	_, err = s3manager.
		NewUploader(c.s3Client, partSizeOpt).
		Upload(ctx, &putObjectInput)

	pb.Wait()

	if errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "\rUpload interrupted by user\n")
		return nil
	}

	return err
}

func (c *Client) EstimatePartSize(f *os.File) (int64, error) {
	size, err := computeSeekerLength(f)
	if err != nil {
		return 0, err
	}

	if size/int64(s3manager.DefaultUploadPartSize) >= int64(s3manager.MaxUploadParts) {
		return (size / int64(s3manager.MaxUploadParts)) + 1, nil
	}

	return s3manager.DefaultUploadPartSize, nil
}

func computeSeekerLength(s io.Seeker) (int64, error) {
	curOffset, err := s.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	endOffset, err := s.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	_, err = s.Seek(curOffset, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return endOffset - curOffset, nil
}

func (c *Client) ShowObject(ctx context.Context, bucket, key string) (output.Outputter, error) {
	object, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object information: %w", err)
	}

	acl, err := c.s3Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve bucket ACL: %w", err)
	}

	out := ShowObjectOutput{
		Path:         key,
		Bucket:       bucket,
		LastModified: object.LastModified.Format(TimestampFormat),
		Size:         object.ContentLength,
		ACL:          ACLFromS3(acl.Grants),
		Metadata:     object.Metadata,
		Headers:      ObjectHeadersFromS3(object),
		URL:          fmt.Sprintf("https://sos-%s.exo.io/%s/%s", c.zone, bucket, key),
	}

	return &out, nil
}

const (
	BucketPrefix    = "sos://"
	TimestampFormat = "2006-01-02 15:04:05 MST"
)

// ObjectHeadersFromS3 returns mutable object headers in a human-friendly
// key/value form.
func ObjectHeadersFromS3(o *s3.GetObjectOutput) map[string]string {
	headers := make(map[string]string)

	if o.CacheControl != nil {
		headers[ObjectHeaderCacheControl] = aws.ToString(o.CacheControl)
	}
	if o.ContentDisposition != nil {
		headers[ObjectHeaderContentDisposition] = aws.ToString(o.ContentDisposition)
	}
	if o.ContentEncoding != nil {
		headers[ObjectHeaderContentEncoding] = aws.ToString(o.ContentEncoding)
	}
	if o.ContentLanguage != nil {
		headers[ObjectHeaderContentLanguage] = aws.ToString(o.ContentLanguage)
	}
	if o.ContentType != nil {
		headers[ObjectHeaderContentType] = aws.ToString(o.ContentType)
	}
	if o.Expires != nil {
		headers[ObjectHeaderExpires] = o.Expires.String()
	}

	return headers
}

type ShowObjectOutput struct {
	Path         string            `json:"name"`
	Bucket       string            `json:"bucket"`
	LastModified string            `json:"last_modified"`
	Size         int64             `json:"size"`
	ACL          ACL               `json:"acl"`
	Metadata     map[string]string `json:"metadata"`
	Headers      map[string]string `json:"headers"`
	URL          string            `json:"url"`
}

func (o *ShowObjectOutput) toJSON() { output.JSON(o) }
func (o *ShowObjectOutput) toText() { output.Text(o) }
func (o *ShowObjectOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Storage"})

	t.Append([]string{"Path", o.Path})
	t.Append([]string{"Bucket", o.Bucket})
	t.Append([]string{"Last Modified", fmt.Sprint(o.LastModified)})
	t.Append([]string{"Size", humanize.IBytes(uint64(o.Size))})
	t.Append([]string{"URL", o.URL})

	t.Append([]string{"ACL", func() string {
		buf := bytes.NewBuffer(nil)
		at := table.NewEmbeddedTable(buf)
		at.SetHeader([]string{" "})
		at.Append([]string{"Read", o.ACL.Read})
		at.Append([]string{"Write", o.ACL.Write})
		at.Append([]string{"Read ACP", o.ACL.ReadACP})
		at.Append([]string{"Write ACP", o.ACL.WriteACP})
		at.Append([]string{"Full Control", o.ACL.FullControl})
		at.Render()

		return buf.String()
	}()})

	t.Append([]string{"Metadata", func() string {
		sortedKeys := func() []string {
			keys := make([]string, 0)
			for k := range o.Metadata {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return keys
		}()

		buf := bytes.NewBuffer(nil)
		at := table.NewEmbeddedTable(buf)
		at.SetHeader([]string{" "})
		for _, k := range sortedKeys {
			at.Append([]string{k, o.Metadata[k]})
		}
		at.Render()

		return buf.String()
	}()})

	t.Append([]string{"Headers", func() string {
		buf := bytes.NewBuffer(nil)
		ht := table.NewEmbeddedTable(buf)
		ht.SetHeader([]string{" "})
		for k, v := range o.Headers {
			ht.Append([]string{k, v})
		}
		ht.Render()

		return buf.String()
	}()})
}
