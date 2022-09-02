package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/table"
)

type storageShowBucketOutput struct {
	Name string            `json:"name"`
	Zone string            `json:"zone"`
	ACL  storageACL        `json:"acl"`
	CORS []storageCORSRule `json:"cors"`
}

func (o *storageShowBucketOutput) toJSON() { outputJSON(o) }
func (o *storageShowBucketOutput) toText() { outputText(o) }
func (o *storageShowBucketOutput) toTable() {
	t := table.NewTable(os.Stdout)
	defer t.Render()
	t.SetHeader([]string{"Storage"})

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Zone", o.Zone})

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

	t.Append([]string{"CORS", func() string {
		buf := bytes.NewBuffer(nil)
		ct := table.NewEmbeddedTable(buf)

		for _, rule := range o.CORS {
			ct.Append([]string{""})
			ct.Append([]string{"{"})
			if rule.AllowedOrigins != nil {
				ct.Append([]string{"", "Allowed Origins", fmt.Sprint(rule.AllowedOrigins)})
			}
			if rule.AllowedMethods != nil {
				ct.Append([]string{"", "Allowed Methods", fmt.Sprint(rule.AllowedMethods)})
			}
			if rule.AllowedHeaders != nil {
				ct.Append([]string{"", "Allowed Headers", fmt.Sprint(rule.AllowedHeaders)})
			}
			ct.Append([]string{"}"})
		}

		ct.Render()

		return buf.String()
	}()})
}

type storageShowObjectOutput struct {
	Path         string            `json:"name"`
	Bucket       string            `json:"bucket"`
	LastModified string            `json:"last_modified"`
	Size         int64             `json:"size"`
	ACL          storageACL        `json:"acl"`
	Metadata     map[string]string `json:"metadata"`
	Headers      map[string]string `json:"headers"`
	URL          string            `json:"url"`
}

func (o *storageShowObjectOutput) toJSON() { outputJSON(o) }
func (o *storageShowObjectOutput) toText() { outputText(o) }
func (o *storageShowObjectOutput) toTable() {
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

func init() {
	storageCmd.AddCommand(&cobra.Command{
		Use:   "show sos://BUCKET/[OBJECT]",
		Short: "Show a bucket/object details",
		Long: fmt.Sprintf(`This command lists Storage buckets and objects.

Supported output template annotations:

	* When showing a bucket: %s
	* When showing an object: %s`,
			strings.Join(outputterTemplateAnnotations(&storageShowBucketOutput{}), ", "),
			strings.Join(outputterTemplateAnnotations(&storageShowObjectOutput{}), ", ")),

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				cmdExitOnUsageError(cmd, "invalid arguments")
			}

			args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				bucket string
				key    string
			)

			parts := strings.SplitN(args[0], "/", 2)
			bucket = parts[0]
			if len(parts) > 1 {
				key = parts[1]
			}

			storage, err := newStorageClient(
				storageClientOptZoneFromBucket(bucket),
			)
			if err != nil {
				return fmt.Errorf("unable to initialize storage client: %w", err)
			}

			if key == "" {
				return output(storage.showBucket(bucket))
			}

			return output(storage.showObject(bucket, key))
		},
	})
}

func (c *storageClient) showBucket(bucket string) (outputter, error) {
	acl, err := c.GetBucketAcl(gContext, &s3.GetBucketAclInput{Bucket: aws.String(bucket)})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve bucket ACL: %w", err)
	}

	cors, err := c.GetBucketCors(gContext, &s3.GetBucketCorsInput{Bucket: aws.String(bucket)})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchCORSConfiguration" {
				cors = &s3.GetBucketCorsOutput{}
			}
		}

		if cors == nil {
			return nil, fmt.Errorf("unable to retrieve bucket CORS configuration: %w", err)
		}
	}

	out := storageShowBucketOutput{
		Name: bucket,
		Zone: c.zone,
		ACL:  storageACLFromS3(acl.Grants),
		CORS: storageCORSRulesFromS3(cors),
	}

	return &out, nil
}

func (c *storageClient) showObject(bucket, key string) (outputter, error) {
	object, err := c.GetObject(gContext, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object information: %w", err)
	}

	acl, err := c.GetObjectAcl(gContext, &s3.GetObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve bucket ACL: %w", err)
	}

	out := storageShowObjectOutput{
		Path:         key,
		Bucket:       bucket,
		LastModified: object.LastModified.Format(storageTimestampFormat),
		Size:         object.ContentLength,
		ACL:          storageACLFromS3(acl.Grants),
		Metadata:     object.Metadata,
		Headers:      storageObjectHeadersFromS3(object),
		URL:          fmt.Sprintf("https://sos-%s.exo.io/%s/%s", c.zone, bucket, key),
	}

	return &out, nil
}
