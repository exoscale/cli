package cmd

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/table"
)

type storageShowBucketOutput struct {
	Name string            `json:"name"`
	Zone string            `json:"zone"`
	ACL  storageACL        `json:"acl"`
	CORS []sos.CORSRule `json:"cors"`
}

func (o *storageShowBucketOutput) toJSON() { output.JSON(o) }
func (o *storageShowBucketOutput) toText() { output.Text(o) }
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

func (o *storageShowObjectOutput) toJSON() { output.JSON(o) }
func (o *storageShowObjectOutput) toText() { output.Text(o) }
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
			strings.Join(output.OutputterTemplateAnnotations(&storageShowBucketOutput{}), ", "),
			strings.Join(output.OutputterTemplateAnnotations(&storageShowObjectOutput{}), ", ")),

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

			storage, err := sos.NewStorageClient(
				storageClientOptZoneFromBucket(bucket),
			)
			if err != nil {
				return fmt.Errorf("unable to initialize storage client: %w", err)
			}

			if key == "" {
				return printOutput(storage.ShowBucket(bucket))
			}

			return printOutput(storage.ShowObject(bucket, key))
		},
	})
}
