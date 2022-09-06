package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	minio "github.com/minio/minio-go/v6"
	"github.com/spf13/cobra"
)

// nolint
const (
	// Canned ACLs
	private                string = "private"
	publicRead             string = "public-read"
	publicReadWrite        string = "public-read-write"
	authenticatedRead      string = "authenticated-read"
	bucketOwnerRead        string = "bucket-owner-read"
	bucketOwnerFullControl string = "bucket-owner-full-control"

	// S3 Grant ACLs response header
	manualRead        string = "X-Amz-Grant-Read"
	manualWrite       string = "X-Amz-Grant-Write"
	manualReadACP     string = "X-Amz-Grant-Read-Acp"
	manualWriteACP    string = "X-Amz-Grant-Write-Acp"
	manualFullControl string = "X-Amz-Grant-Full-Control"

	// S3 Grant ACLs response body
	sosACLRead        string = "READ"
	sosACLWrite       string = "WRITE"
	sosACLReadACP     string = "READ_ACP"
	sosACLWriteACP    string = "WRITE_ACP"
	sosACLFullControl string = "FULL_CONTROL"
)

var sosACLCmd = &cobra.Command{
	Use:   "acl",
	Short: "Object(s) ACLs management",
}

func init() {
	sosCmd.AddCommand(sosACLCmd)
}

var sosAddACLCmd = &cobra.Command{
	Use:   "add BUCKET OBJECT|PREFIX",
	Short: "Add ACL(s) to objects",
	Long: `This commands adds ACL(s) to objects in a bucket. It is possible to
set ACLs either on a single object, or recursively from a prefix
using the flag "--recursive". To recurse across the whole bucket,
specify "/" as prefix.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		path := args[1]

		// The "/" path is just a trick for CLI users to signify they mean the bucket root:
		// for SOS the actual bucket root is an empty prefix (i.e. ""), however passing
		// an empty string as CLI option value has a different meaning.
		if path == "/" {
			path = ""
		}

		meta, err := getACL(cmd)
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		if meta == nil {
			println("error: You have to choose one flag")
			if err = cmd.Usage(); err != nil {
				return err
			}
			return fmt.Errorf("error: You have to choose one flag")
		}

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		for obj := range sosClient.ListObjects(bucket, path, recursive, gContext.Done()) {
			// Skip "folders", as they don't support ACLs.
			if strings.HasSuffix(obj.Key, "/") {
				continue
			}

			objInfo, err := sosClient.GetObjectACLWithContext(gContext, bucket, obj.Key)
			if err != nil {
				return err
			}

			objInfo.Metadata.Add("content-type", objInfo.ContentType)

			src := minio.NewSourceInfo(bucket, objInfo.Key, nil)

			// When the Object acl is updated from Canned ACL(X-Amz-Acl) to
			// Grant ACL(X-Amz-Grant), we have to remove Canned ACL before.
			_, hasNewCannedACL := meta["X-Amz-Acl"]
			_, hasCannedACL := objInfo.Metadata["X-Amz-Acl"]
			if hasCannedACL && !hasNewCannedACL {
				// Remove Canned ACL from the header to let Grant ACL take effect.
				objInfo.Metadata.Del("X-Amz-Acl")
				// This lets the objInfo full control grantee to keep the control on the objInfo,
				// if the flag "--full-control" is not specified.
				var fullControl string
				for _, g := range objInfo.Grant {
					if g.Permission == sosACLFullControl {
						fullControl = g.Grantee.ID
					}
				}
				if fullControl == "" {
					return fmt.Errorf(`objInfo %q has no "FULL_CONTROL" grantee`, objInfo)
				}
				objInfo.Metadata.Add(manualFullControl, "id="+fullControl)
			}

			mergeHeader(src.Headers, objInfo.Metadata)

			// Destination objInfo
			dst, err := minio.NewDestinationInfo(bucket, objInfo.Key, nil, meta)
			if err != nil {
				return err
			}

			// Copy objInfo call
			err = sosClient.CopyObject(dst, src)
			if err != nil {
				return err
			}

			if !gQuiet && !recursive {
				acl, err := getDefaultCannedACL(cmd)
				if err != nil {
					return err
				}

				if acl == publicReadWrite || acl == publicRead {
					fmt.Printf("https://sos-%s.exo.io/%s/%s\n", location, bucket, objInfo.Key)
				}
			}
		}

		return nil
	},
}

// merge src header in dst header
func mergeHeader(dst, src http.Header) {
	for k, v := range src {
		dst[k] = v
	}
}

func getACL(cmd *cobra.Command) (map[string]string, error) {
	meta := map[string]string{}

	defACL, err := getDefaultCannedACL(cmd)
	if err != nil {
		return nil, err
	}

	if defACL != "" {
		meta["X-Amz-Acl"] = defACL
		return meta, nil
	}

	manualACLs, err := getManualACL(cmd)
	if err != nil {
		return nil, err
	}

	if manualACLs == nil {
		return nil, nil
	}

	for k, v := range manualACLs {
		for i := range v {
			v[i] = fmt.Sprintf("id=%s", v[i])
		}

		meta[k] = strings.Join(v, ", ")
	}

	return meta, nil
}

func getDefaultCannedACL(cmd *cobra.Command) (string, error) {
	acl, err := cmd.Flags().GetBool(private)
	if err != nil {
		return "", err
	}
	if acl {
		return private, nil
	}

	acl, err = cmd.Flags().GetBool(publicRead)
	if err != nil {
		return "", err
	}
	if acl {
		return publicRead, nil
	}

	acl, err = cmd.Flags().GetBool(publicReadWrite)
	if err != nil {
		return "", err
	}
	if acl {
		return publicReadWrite, nil
	}

	acl, err = cmd.Flags().GetBool(authenticatedRead)
	if err != nil {
		return "", err
	}
	if acl {
		return authenticatedRead, nil
	}

	acl, err = cmd.Flags().GetBool(bucketOwnerRead)
	if err != nil {
		return "", err
	}
	if acl {
		return bucketOwnerRead, nil
	}

	acl, err = cmd.Flags().GetBool(bucketOwnerFullControl)
	if err != nil {
		return "", err
	}
	if acl {
		return bucketOwnerFullControl, nil
	}

	return "", nil
}

func getManualACL(cmd *cobra.Command) (map[string][]string, error) {
	res := map[string][]string{}

	acl, err := cmd.Flags().GetString("read")
	if err != nil {
		return nil, err
	}
	if acl != "" {
		res[manualRead] = getCommaflag(acl)
	}

	acl, err = cmd.Flags().GetString("write")
	if err != nil {
		return nil, err
	}
	if acl != "" {
		res[manualWrite] = getCommaflag(acl)
	}

	acl, err = cmd.Flags().GetString("read-acp")
	if err != nil {
		return nil, err
	}
	if acl != "" {
		res[manualReadACP] = getCommaflag(acl)
	}

	acl, err = cmd.Flags().GetString("write-acp")
	if err != nil {
		return nil, err
	}
	if acl != "" {
		res[manualWriteACP] = getCommaflag(acl)
	}

	acl, err = cmd.Flags().GetString("full-control")
	if err != nil {
		return nil, err
	}
	if acl != "" {
		res[manualFullControl] = getCommaflag(acl)
	}

	if len(res) == 0 {
		return nil, nil
	}

	return res, err
}

func init() {
	sosACLCmd.AddCommand(sosAddACLCmd)
	sosAddACLCmd.Flags().SortFlags = false
	sosAddACLCmd.Flags().Bool("recursive", false, "Set ACL recursively")

	// Canned ACLs
	sosAddACLCmd.Flags().BoolP(private, "p", false, "Canned ACL private")
	sosAddACLCmd.Flags().BoolP(publicRead, "r", false, "Canned ACL public read")
	sosAddACLCmd.Flags().BoolP(publicReadWrite, "w", false, "Canned ACL public read and write")
	sosAddACLCmd.Flags().BoolP(authenticatedRead, "", false, "Canned ACL authenticated read")
	sosAddACLCmd.Flags().BoolP(bucketOwnerRead, "", false, "Canned ACL bucket owner read")
	sosAddACLCmd.Flags().BoolP(bucketOwnerFullControl, "f", false, "Canned ACL bucket owner full control")

	// Manual ACLs
	sosAddACLCmd.Flags().StringP("read", "", "", "Manual acl edit grant read e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("write", "", "", "Manual acl edit grant write e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("read-acp", "", "", "Manual acl edit grant acp read e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("write-acp", "", "", "Manual acl edit grant acp write e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("full-control", "", "", "Manual acl edit grant full control e.g(value, value, ...)")
}

var sosRemoveACLCmd = &cobra.Command{
	Use:     "remove BUCKET OBJECT",
	Short:   "Remove ACL(s) from an object",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

		meta, err := getManualACLBool(cmd)
		if err != nil {
			return err
		}

		if meta == nil {
			println("error: You have to choose one flag")
			if err = cmd.Usage(); err != nil {
				return err
			}
			return fmt.Errorf("error: You have to choose one flag")
		}

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		objInfo, err := sosClient.GetObjectACLWithContext(gContext, bucket, object)
		if err != nil {
			return err
		}

		objInfo.Metadata.Add("content-type", objInfo.ContentType)
		objInfo.Metadata.Add("x-amz-metadata-directive", "REPLACE")

		src := minio.NewSourceInfo(bucket, object, nil)

		_, okHeader := objInfo.Metadata["X-Amz-Acl"]

		if okHeader {
			return fmt.Errorf("error: No Manual ACL are set")
		}

		for _, k := range meta {
			objInfo.Metadata.Del(k)
		}

		mergeHeader(src.Headers, objInfo.Metadata)

		// Destination object
		dst, err := minio.NewDestinationInfo(bucket, object, nil, nil)
		if err != nil {
			return err
		}

		// Copy object call
		return sosClient.CopyObject(dst, src)
	},
}

func init() {
	sosACLCmd.AddCommand(sosRemoveACLCmd)
	sosRemoveACLCmd.Flags().SortFlags = false
	sosRemoveACLCmd.Flags().BoolP("read", "r", false, "Remove grant read ACL")
	sosRemoveACLCmd.Flags().BoolP("write", "w", false, "Remove grant write ACL")
	sosRemoveACLCmd.Flags().BoolP("read-acp", "", false, "Remove grant acp read ACL")
	sosRemoveACLCmd.Flags().BoolP("write-acp", "", false, "Remove grant acp write ACL")
	sosRemoveACLCmd.Flags().BoolP("full-control", "f", false, "Remove grant full control ACL")
}

func getManualACLBool(cmd *cobra.Command) ([]string, error) {
	var res []string

	acl, err := cmd.Flags().GetBool("read")
	if err != nil {
		return nil, err
	}
	if acl {
		res = append(res, manualRead)
	}

	acl, err = cmd.Flags().GetBool("write")
	if err != nil {
		return nil, err
	}
	if acl {
		res = append(res, manualWrite)
	}

	acl, err = cmd.Flags().GetBool("read-acp")
	if err != nil {
		return nil, err
	}
	if acl {
		res = append(res, manualReadACP)
	}

	acl, err = cmd.Flags().GetBool("write-acp")
	if err != nil {
		return nil, err
	}
	if acl {
		res = append(res, manualWriteACP)
	}

	acl, err = cmd.Flags().GetBool("full-control")
	if err != nil {
		return nil, err
	}
	if acl {
		res = append(res, manualFullControl)
	}

	return res, nil
}

var sosShowACLCmd = &cobra.Command{
	Use:     "list BUCKET OBJECT",
	Short:   "list Object ACLs",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

		sosClient, err := newSOSClient()
		if err != nil {
			return err
		}

		location, err := sosClient.GetBucketLocation(bucket)
		if err != nil {
			return err
		}

		if err := sosClient.setZone(location); err != nil {
			return err
		}

		objInfo, err := sosClient.GetObjectACLWithContext(gContext, bucket, object)
		if err != nil {
			return err
		}

		cannedACL, okHeader := objInfo.Metadata["X-Amz-Acl"]

		table := table.NewTable(os.Stdout)
		table.SetHeader([]string{"File Name", "ACL", "Value"})

		if okHeader && len(cannedACL) > 0 {
			table.Append([]string{objInfo.Key, "Canned", cannedACL[0]})
		} else {
			for _, g := range objInfo.Grant {
				table.Append([]string{objInfo.Key, formatGrantKey(g.Permission), g.Grantee.DisplayName})
			}
		}

		table.Render()

		return nil
	},
}

func formatGrantKey(k string) string {
	var res string

	switch k {
	case sosACLRead:
		res = "Read"
	case sosACLWrite:
		res = "Write"
	case sosACLReadACP:
		res = "Read ACP"
	case sosACLWriteACP:
		res = "Write ACP"
	case sosACLFullControl:
		res = "Full Control"
	}

	return res
}

func init() {
	sosACLCmd.AddCommand(sosShowACLCmd)
}
