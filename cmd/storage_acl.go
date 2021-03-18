package cmd

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type storageACL struct {
	Canned      string `json:"-"`
	Read        string `json:"read"`
	Write       string `json:"write"`
	ReadACP     string `json:"read_acp"`
	WriteACP    string `json:"write_acp"`
	FullControl string `json:"full_control"`
}

const (
	storageACLGranteeAllUsers           = "http://acs.amazonaws.com/groups/global/AllUsers"
	storageACLGranteeAuthenticatedUsers = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"

	storageSetACLCmdFlagRead        = "read"
	storageSetACLCmdFlagWrite       = "write"
	storageSetACLCmdFlagReadACP     = "read-acp"
	storageSetACLCmdFlagWriteACP    = "write-acp"
	storageSetACLCmdFlagFullControl = "full-control"
)

// toS3Grants converts the local ACL representation to the s3types.Grant format.
func (a storageACL) toS3Grants() []s3types.Grant {
	grants := make([]s3types.Grant, 0)

	if a.Read != "" {
		grants = append(grants, s3types.Grant{
			Grantee:    storageACLGranteeToS3(a.Read),
			Permission: s3types.PermissionRead,
		})
	}
	if a.Write != "" {
		grants = append(grants, s3types.Grant{
			Grantee:    storageACLGranteeToS3(a.Write),
			Permission: s3types.PermissionWrite,
		})
	}
	if a.ReadACP != "" {
		grants = append(grants, s3types.Grant{
			Grantee:    storageACLGranteeToS3(a.ReadACP),
			Permission: s3types.PermissionReadAcp,
		})
	}
	if a.WriteACP != "" {
		grants = append(grants, s3types.Grant{
			Grantee:    storageACLGranteeToS3(a.WriteACP),
			Permission: s3types.PermissionWriteAcp,
		})
	}
	if a.FullControl != "" {
		grants = append(grants, s3types.Grant{
			Grantee:    storageACLGranteeToS3(a.FullControl),
			Permission: s3types.PermissionFullControl,
		})
	}

	return grants
}

// storageACLFromS3 converts an S3 ACL Grant to the local ACL representation.
func storageACLFromS3(grants []s3types.Grant) storageACL {
	acl := storageACL{
		Read:        "-",
		Write:       "-",
		ReadACP:     "-",
		WriteACP:    "-",
		FullControl: "-",
	}

	for _, grant := range grants {
		switch grant.Permission {
		case s3types.PermissionRead:
			acl.Read = storageACLGranteeFromS3(grant.Grantee)

		case s3types.PermissionWrite:
			acl.Write = storageACLGranteeFromS3(grant.Grantee)

		case s3types.PermissionReadAcp:
			acl.ReadACP = storageACLGranteeFromS3(grant.Grantee)

		case s3types.PermissionWriteAcp:
			acl.WriteACP = storageACLGranteeFromS3(grant.Grantee)

		case s3types.PermissionFullControl:
			acl.FullControl = storageACLGranteeFromS3(grant.Grantee)
		}
	}

	return acl
}

// storageACLGranteeFromS3 returns a human-friendly representation of an S3 ACL
// Grantee.
func storageACLGranteeFromS3(v *s3types.Grantee) string {
	if v.Type == s3types.TypeCanonicalUser {
		return aws.ToString(v.DisplayName)
	}

	switch aws.ToString(v.URI) {
	case storageACLGranteeAllUsers:
		return "ALL_USERS"

	case storageACLGranteeAuthenticatedUsers:
		return "AUTHENTICATED_USERS"
	}

	return "-"
}

// storageACLGranteeToS3 converts a human-friendly ACL grantee representation
// to the S3 format.
func storageACLGranteeToS3(v string) *s3types.Grantee {
	switch v {
	case "ALL_USERS":
		return &s3types.Grantee{
			Type: s3types.TypeGroup,
			URI:  aws.String(storageACLGranteeAllUsers),
		}

	case "AUTHENTICATED_USERS":
		return &s3types.Grantee{
			Type: s3types.TypeGroup,
			URI:  aws.String(storageACLGranteeAuthenticatedUsers),
		}

	default:
		return &s3types.Grantee{
			Type: s3types.TypeCanonicalUser,
			ID:   aws.String(v),
		}
	}
}

// storageACLToCopyObject updates the object to be copied with S3 ACL information.
func storageACLToCopyObject(acl *s3.GetObjectAclOutput, o *s3.CopyObjectInput) {
	s3GranteeToString := func(g *s3types.Grantee) *string {
		if g.Type == s3types.TypeCanonicalUser {
			return aws.String("id=" + aws.ToString(g.ID))
		}

		switch aws.ToString(g.URI) {
		case storageACLGranteeAllUsers:
			return aws.String("uri=" + storageACLGranteeAllUsers)

		case storageACLGranteeAuthenticatedUsers:
			return aws.String("uri=" + storageACLGranteeAuthenticatedUsers)
		}

		return nil
	}

	o.GrantFullControl = aws.String("id=" + aws.ToString(acl.Owner.ID))

	for _, grant := range acl.Grants {
		switch grant.Permission {
		case s3types.PermissionRead:
			o.GrantRead = s3GranteeToString(grant.Grantee)

		// Write permission is not supported on S3 objects:
		// https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#permissions

		case s3types.PermissionReadAcp:
			o.GrantReadACP = s3GranteeToString(grant.Grantee)

		case s3types.PermissionWriteAcp:
			o.GrantWriteACP = s3GranteeToString(grant.Grantee)

		case s3types.PermissionFullControl:
			o.GrantFullControl = s3GranteeToString(grant.Grantee)
		}
	}
}

func s3BucketCannedACLToStrings() []string {
	s3BucketCannedACLs := s3types.BucketCannedACLPrivate.Values()

	list := make([]string, len(s3BucketCannedACLs))
	for i, v := range s3BucketCannedACLs {
		list[i] = string(v)
	}

	return list
}

func s3ObjectCannedACLToStrings() []string {
	s3ObjectCannedACLs := s3types.ObjectCannedACLPrivate.Values()

	list := make([]string, len(s3ObjectCannedACLs))
	for i, v := range s3ObjectCannedACLs {
		list[i] = string(v)
	}

	return list
}
