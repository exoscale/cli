package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

type sosACLShowOutput struct {
	Access string `json:"access"`
	Value  string `json:"value"`
}

type sosMetadataShowOutput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type sosHeadersShowOutput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type sosShowOutput struct {
	Object   string                  `json:"object"`
	URL      string                  `json:"url"`
	ACL      []sosACLShowOutput      `json:"acl"`
	Metadata []sosMetadataShowOutput `json:"metadata"`
	Headers  []sosHeadersShowOutput  `json:"headers"`
}

func (o *sosShowOutput) toJSON() { outputJSON(o) }

func (o *sosShowOutput) toText() { outputText(o) }

func (o *sosShowOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{o.Object})

	if o.ACL != nil {
		buf := bytes.NewBuffer(nil)
		at := table.NewEmbeddedTable(buf)
		at.SetHeader([]string{"access", "value"})
		for _, a := range o.ACL {
			at.Append([]string{a.Access, a.Value})
		}
		at.Render()
		t.Append([]string{"ACL", buf.String()})
	}

	if o.Metadata != nil {
		buf := bytes.NewBuffer(nil)
		mt := table.NewEmbeddedTable(buf)
		mt.SetHeader([]string{"key", "value"})
		for _, m := range o.Metadata {
			mt.Append([]string{m.Key, m.Value})
		}
		mt.Render()
		t.Append([]string{"Metadata", buf.String()})
	}

	if o.Headers != nil {
		buf := bytes.NewBuffer(nil)
		ht := table.NewEmbeddedTable(buf)
		ht.SetHeader([]string{"key", "value"})
		for _, h := range o.Headers {
			ht.Append([]string{h.Key, h.Value})
		}
		ht.Render()
		t.Append([]string{"Headers", buf.String()})
	}

	t.Append([]string{"URL", fmt.Sprint(o.URL)})

	t.Render()
}

// sosShowCmd represents the show command
var sosShowCmd = &cobra.Command{
	Use:     "show <bucket name> <oject name>",
	Short:   "show file and folder",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		return output(showSOS(args[0], args[1], cmd))
	},
}

func showSOS(bucket, object string, cmd *cobra.Command) (outputter, error) {
	certsFile, err := cmd.Parent().Flags().GetString("certs-file")
	if err != nil {
		return nil, err
	}

	sosClient, err := newSOSClient(certsFile)
	if err != nil {
		return nil, err
	}

	location, err := sosClient.GetBucketLocation(bucket)
	if err != nil {
		return nil, err
	}

	if err := sosClient.setZone(location); err != nil {
		return nil, err
	}

	objInfo, err := sosClient.GetObjectACL(bucket, object)
	if err != nil {
		return nil, err
	}

	cannedACL, okHeader := objInfo.Metadata["X-Amz-Acl"]

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"ACL", "METADATA", "HEADER"})

	var (
		acls      []sosACLShowOutput
		metadatas []sosMetadataShowOutput
		headers   []sosHeadersShowOutput
	)

	if okHeader && len(cannedACL) > 0 {
		acls = append(acls, sosACLShowOutput{
			Access: "Canned",
			Value:  cannedACL[0],
		})
	}

	if objInfo.ContentType != "" {
		headers = append(headers, sosHeadersShowOutput{
			Key:   "content-type",
			Value: objInfo.ContentType,
		})
	}

	for k, v := range objInfo.Metadata {
		if len(v) > 0 {
			if isGrantACL(k) {
				s := getGrantValue(v)
				acls = append(acls, sosACLShowOutput{
					Access: formatGrantKey(k),
					Value:  s,
				})
			}

			k = strings.ToLower(k)

			if strings.HasPrefix(k, "x-amz-meta-") && len(v) > 0 {
				metadatas = append(metadatas, sosMetadataShowOutput{
					Key:   k[len("x-amz-meta-"):],
					Value: v[0],
				})
			}

			if isStandardHeader(k) && len(v) > 0 {
				headers = append(headers, sosHeadersShowOutput{
					Key:   k,
					Value: v[0],
				})
			}
		}
	}

	out := sosShowOutput{
		Object:   object,
		URL:      fmt.Sprintf("https://sos-%s.exo.io/%s/%s", location, bucket, object),
		ACL:      acls,
		Metadata: metadatas,
		Headers:  headers,
	}

	return &out, nil
}

func init() {
	sosCmd.AddCommand(sosShowCmd)
}
