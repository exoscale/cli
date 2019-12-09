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
	//Canned ACLs
	private                string = "private"
	publicRead             string = "public-read"
	publicReadWrite        string = "public-read-write"
	authenticatedRead      string = "authenticated-read"
	bucketOwnerRead        string = "bucket-owner-read"
	bucketOwnerFullControl string = "bucket-owner-full-control"

	//S3 Grant ACLs response header
	manualRead        string = "X-Amz-Grant-Read"
	manualWrite       string = "X-Amz-Grant-Write"
	manualReadACP     string = "X-Amz-Grant-Read-Acp"
	manualWriteACP    string = "X-Amz-Grant-Write-Acp"
	manualFullControl string = "X-Amz-Grant-Full-Control"

	//S3 Grant ACLs response body
	sosACLRead        string = "READ"
	sosACLWrite       string = "WRITE"
	sosACLReadACP     string = "READ_ACP"
	sosACLWriteACP    string = "WRITE_ACP"
	sosACLFullControl string = "FULL_CONTROL"
)

// aclCmd represents the acl command
var sosACLCmd = &cobra.Command{
	Use:   "acl <bucket name> <object name> [object name] ...",
	Short: "Object(s) ACLs management",
}

func init() {
	sosCmd.AddCommand(sosACLCmd)
}

// aclCmd represents the acl command
var sosAddACLCmd = &cobra.Command{
	Use:   "add <bucket name> <object name> [object name] ...",
	Short: "Add ACL(s) to an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

		meta, err := getACL(cmd)
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

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
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

		objInfo, err := sosClient.GetObjectACL(bucket, object)
		if err != nil {
			return err
		}

		objInfo.Metadata.Add("content-type", objInfo.ContentType)

		src := minio.NewSourceInfo(bucket, object, nil)

		_, okMeta := meta["X-Amz-Acl"]
		_, okHeader := objInfo.Metadata["X-Amz-Acl"]

		if okHeader && !okMeta {
			objInfo.Metadata.Del("X-Amz-Acl")
			objInfo.Metadata.Add(manualFullControl, "id="+gCurrentAccount.AccountName())
		}

		mergeHeader(src.Headers, objInfo.Metadata)

		// Destination object
		dst, err := minio.NewDestinationInfo(bucket, object, nil, meta)
		if err != nil {
			return err
		}

		// Copy object call
		return sosClient.CopyObject(dst, src)
	},
}

//merge src header in dst header
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

	//Canned ACLs
	sosAddACLCmd.Flags().BoolP(private, "p", false, "Canned ACL private")
	sosAddACLCmd.Flags().BoolP(publicRead, "r", false, "Canned ACL public read")
	sosAddACLCmd.Flags().BoolP(publicReadWrite, "w", false, "Canned ACL public read and write")
	sosAddACLCmd.Flags().BoolP(authenticatedRead, "", false, "Canned ACL authenticated read")
	sosAddACLCmd.Flags().BoolP(bucketOwnerRead, "", false, "Canned ACL bucket owner read")
	sosAddACLCmd.Flags().BoolP(bucketOwnerFullControl, "f", false, "Canned ACL bucket owner full control")

	//Manual ACLs
	sosAddACLCmd.Flags().StringP("read", "", "", "Manual acl edit grant read e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("write", "", "", "Manual acl edit grant write e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("read-acp", "", "", "Manual acl edit grant acp read e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("write-acp", "", "", "Manual acl edit grant acp write e.g(value, value, ...)")
	sosAddACLCmd.Flags().StringP("full-control", "", "", "Manual acl edit grant full control e.g(value, value, ...)")
}

// aclCmd represents the acl command
var sosRemoveACLCmd = &cobra.Command{
	Use:     "remove <bucket name> <object name> [object name] ...",
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

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
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

		objInfo, err := sosClient.GetObjectACL(bucket, object)
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

// aclCmd represents the acl command
var sosShowACLCmd = &cobra.Command{
	Use:     "list <bucket name> <object name>",
	Short:   "list Object ACLs",
	Aliases: gListAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}
		bucket := args[0]
		object := args[1]

		certsFile, err := cmd.Parent().Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		sosClient, err := newSOSClient(certsFile)
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

		objInfo, err := sosClient.GetObjectACL(bucket, object)
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
