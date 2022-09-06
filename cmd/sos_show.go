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
	URL      string                  `json:"url"`
	ACL      []sosACLShowOutput      `json:"acl"`
	Metadata []sosMetadataShowOutput `json:"metadata"`
	Headers  []sosHeadersShowOutput  `json:"headers"`
}

func (o *sosShowOutput) toJSON() { outputJSON(o) }

func (o *sosShowOutput) toText() { outputText(o) }

func (o *sosShowOutput) toTable() {
	t := table.NewTable(os.Stdout)

	if o.ACL != nil {
		buf := bytes.NewBuffer(nil)
		at := table.NewEmbeddedTable(buf)

		switch {
		case o.ACL[0].Access == "Canned":
			at.Append([]string{o.ACL[0].Value})
		default:
			at.SetHeader([]string{"access", "value"})

			for _, a := range o.ACL {
				at.Append([]string{a.Access, a.Value})
			}
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

	t.Append([]string{"URL", o.URL})

	t.Render()
}

var sosShowCmd = &cobra.Command{
	Use:     "show BUCKET OBJECT",
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
	sosClient, err := newSOSClient()
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

	var (
		acls     []sosACLShowOutput
		metadata []sosMetadataShowOutput
		headers  []sosHeadersShowOutput
	)

	if okHeader && len(cannedACL) > 0 {
		acls = append(acls, sosACLShowOutput{
			Access: "Canned",
			Value:  cannedACL[0],
		})
	} else {
		for _, g := range objInfo.Grant {
			acls = append(acls, sosACLShowOutput{
				Access: formatGrantKey(g.Permission),
				Value:  g.Grantee.DisplayName,
			})
		}
	}

	if objInfo.ContentType != "" {
		headers = append(headers, sosHeadersShowOutput{
			Key:   "content-type",
			Value: objInfo.ContentType,
		})
	}

	for k, v := range objInfo.Metadata {
		k = strings.ToLower(k)

		if strings.HasPrefix(k, "x-amz-meta-") && len(v) > 0 {
			metadata = append(metadata, sosMetadataShowOutput{
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

	out := sosShowOutput{
		URL:      fmt.Sprintf("https://sos-%s.exo.io/%s/%s", location, bucket, object),
		ACL:      acls,
		Metadata: metadata,
		Headers:  headers,
	}

	return &out, nil
}

func init() {
	sosCmd.AddCommand(sosShowCmd)
}
