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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dustin/go-humanize"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos/object"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
)

func (c *Client) DeleteObjects(ctx context.Context, bucket, prefix string, recursive bool) ([]types.DeletedObject, error) {
	deleteList := make([]types.ObjectIdentifier, 0)
	err := c.ForEachObject(ctx, bucket, prefix, recursive, func(o object.ObjectInterface) error {
		deleteList = append(deleteList, types.ObjectIdentifier{Key: o.GetKey()})
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing objects to delete: %w", err)
	}

	// The S3 DeleteObjects API call is limited to 1000 keys per call, as a
	// precaution we're batching deletes.
	maxKeys := 1000
	deleted := make([]types.DeletedObject, 0)

	for i := 0; i < len(deleteList); i += maxKeys {
		j := i + maxKeys
		if j > len(deleteList) {
			j = len(deleteList)
		}

		res, err := c.S3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: &bucket,
			Delete: &types.Delete{Objects: deleteList[i:j]},
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

	// TODO(sauterp) is there a safer way to achieve this?
	psClient := s3.NewPresignClient(c.S3Client.(*s3.Client), func(o *s3.PresignOptions) {
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
	Bucket      string
	Prefix      string
	Source      string
	Destination string
	Objects     []*types.Object
	Recursive   bool
	Overwrite   bool
	DryRun      bool
}

func (c *Client) DownloadFiles(ctx context.Context, config *DownloadConfig) error {
	if len(config.Objects) > 1 && !strings.HasSuffix(config.Destination, "/") {
		return errors.New(`multiple files to download, destination must end with "/"`)
	}

	// Handle relative filesystem destination (e.g. ".", "../.." etc.)
	if dstInfo, err := os.Stat(config.Destination); err == nil {
		if dstInfo.IsDir() && !strings.HasSuffix(config.Destination, "/") {
			config.Destination += "/"
		}
	}

	if config.DryRun {
		fmt.Println("[DRY-RUN]")
	}

	for _, object := range config.Objects {
		dst := func() string {
			if strings.HasSuffix(config.Source, "/") {
				return path.Join(config.Destination, strings.TrimPrefix(aws.ToString(object.Key), config.Prefix))
			}

			if strings.HasSuffix(config.Destination, "/") {
				return path.Join(config.Destination, path.Base(aws.ToString(object.Key)))
			}

			return path.Join(config.Destination)
		}()

		if config.DryRun {
			fmt.Printf("%s/%s -> %s\n", config.Bucket, aws.ToString(object.Key), dst)
			continue
		}

		if _, err := os.Stat(dst); err == nil && !config.Overwrite {
			return fmt.Errorf("file %q already exists, use flag `-f` to overwrite", dst)
		}

		if _, err := os.Stat(path.Dir(dst)); errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(path.Dir(dst), 0o755); err != nil {
				return err
			}
		}

		if err := c.DownloadFile(ctx, config.Bucket, object, dst); err != nil {
			return err
		}
	}

	return nil
}

// proxyWriterAt is a variant of the internal mpb.proxyWriterTo struct,
// required for using mpb with s3manager batch download manager (accepting
// a io.WriterAt interface) since mpb.Bar's ProxyReader() method only
// supports io.Reader and io.WriterTo interfaces.
type proxyWriterAt struct {
	wt  io.WriterAt
	bar *mpb.Bar
	iT  time.Time
}

func (prox *proxyWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = prox.wt.WriteAt(p, off)
	if n > 0 {
		prox.bar.IncrInt64(int64(n), time.Since(prox.iT))
		prox.iT = time.Now()
	}

	if err == io.EOF {
		go prox.bar.SetTotal(0, true)
	}

	return n, err
}

func (c *Client) DownloadFile(ctx context.Context, bucket string, object *types.Object, dst string) error {
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
		NewDownloader(c.S3Client).
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

type ListBucketsOutput []ListBucketsItemOutput

func (o *ListBucketsOutput) ToJSON() { output.JSON(o) }
func (o *ListBucketsOutput) ToText() { output.Text(o) }
func (o *ListBucketsOutput) ToTable() {
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

func GetCommonPrefixDeduplicator(stream bool) func([]types.CommonPrefix) []string {
	dirs := make(map[string]struct{})

	return func(prefixes []types.CommonPrefix) []string {
		var deduplicatedPrefixes []string

		for _, cp := range prefixes {
			dir := aws.ToString(cp.Prefix)
			if _, ok := dirs[dir]; !ok {
				if stream {
					fmt.Println(dir)
				} else {
					deduplicatedPrefixes = append(deduplicatedPrefixes, dir)
				}
				dirs[dir] = struct{}{}
			}
		}

		return deduplicatedPrefixes
	}
}

type StorageUploadConfig struct {
	Bucket    string
	Prefix    string
	ACL       string
	Recursive bool
	DryRun    bool
}

func (c *Client) UploadFiles(ctx context.Context, sources []string, config *StorageUploadConfig) error {
	if len(sources) > 1 && !strings.HasSuffix(config.Prefix, "/") {
		return errors.New(`multiple files to upload, destination must end with "/"`)
	}

	if config.DryRun {
		fmt.Println("[DRY-RUN]")
	}

	for _, src := range sources {
		src := src

		srcInfo, err := os.Stat(src)
		if err != nil {
			return err
		}

		if srcInfo.IsDir() {
			if !config.Recursive {
				return fmt.Errorf("%q is a directory, use flag `-r` to upload recursively", src)
			}

			err = filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
				var (
					key    string
					prefix = config.Prefix
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

				if config.DryRun {
					fmt.Printf("%s -> %s/%s\n", src, config.Bucket, key)
					return nil
				}

				return c.UploadFile(ctx, config.Bucket, filePath, key, config.ACL)
			})
			if err != nil {
				return err
			}
		} else {
			key := path.Base(src)

			if prefix := config.Prefix; prefix != "" {
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

			if config.DryRun {
				fmt.Printf("%s -> %s/%s\n", src, config.Bucket, key)
				continue
			}

			if err := c.UploadFile(ctx, config.Bucket, src, key, config.ACL); err != nil {
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
		putObjectInput.ACL = types.ObjectCannedACL(acl)
	}

	_, err = c.NewUploader(c.S3Client, partSizeOpt).
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

func (c *Client) ShowObject(ctx context.Context, bucket, key, versionID string) (*ShowObjectOutput, error) {
	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if versionID != "" {
		fmt.Println("setting version id", versionID)
		getObjInput.VersionId = aws.String(versionID)
	}

	obj, err := c.S3Client.GetObject(ctx, getObjInput)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object information: %w", err)
	}

	acl, err := c.S3Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve bucket ACL: %w", err)
	}

	// TODO is it possible to show the version number
	out := ShowObjectOutput{
		Path:         key,
		Bucket:       bucket,
		LastModified: obj.LastModified.Format(object.TimestampFormat),
		Size:         obj.ContentLength,
		ACL:          ACLFromS3(acl.Grants),
		Metadata:     obj.Metadata,
		Headers:      ObjectHeadersFromS3(obj),
		URL:          fmt.Sprintf("https://sos-%s.exo.io/%s/%s", c.Zone, bucket, key),
	}

	return &out, nil
}

const (
	BucketPrefix = "sos://"
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

func (o *ShowObjectOutput) ToJSON() { output.JSON(o) }
func (o *ShowObjectOutput) ToText() { output.Text(o) }
func (o *ShowObjectOutput) ToTable() {
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
