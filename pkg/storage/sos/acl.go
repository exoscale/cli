package sos

import (
	"context"
	"fmt"
	"strings"

	"github.com/exoscale/cli/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	ACLGranteeAllUsers           = "http://acs.amazonaws.com/groups/global/AllUsers"
	ACLGranteeAuthenticatedUsers = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"

	SetACLCmdFlagRead        = "read"
	SetACLCmdFlagWrite       = "write"
	SetACLCmdFlagReadACP     = "read-acp"
	SetACLCmdFlagWriteACP    = "write-acp"
	SetACLCmdFlagFullControl = "full-control"
)

type ACL struct {
	Canned      string `json:"-"`
	Read        string `json:"read"`
	Write       string `json:"write"`
	ReadACP     string `json:"read_acp"`
	WriteACP    string `json:"write_acp"`
	FullControl string `json:"full_control"`
}

func (c *Client) SetBucketACL(ctx context.Context, bucket string, acl *ACL) error {
	s3ACL := s3.PutBucketAclInput{Bucket: aws.String(bucket)}

	if acl.Canned != "" {
		if !utils.IsInList(BucketCannedACLToStrings(), acl.Canned) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl.Canned,
				strings.Join(BucketCannedACLToStrings(), ", "))
		}

		s3ACL.ACL = s3types.BucketCannedACL(acl.Canned)
	} else {
		s3ACL.AccessControlPolicy = &s3types.AccessControlPolicy{Grants: acl.toS3Grants()}

		// As a safety precaution, if the caller didn't explicitly set a Grantee
		// with the FULL_CONTROL permission we set it to the current bucket owner.
		if acl.FullControl == "" {
			curACL, err := c.S3Client.GetBucketAcl(ctx, &s3.GetBucketAclInput{Bucket: aws.String(bucket)})
			if err != nil {
				return fmt.Errorf("unable to retrieve current ACL: %w", err)
			}

			s3ACL.AccessControlPolicy.Grants = append(s3ACL.AccessControlPolicy.Grants, s3types.Grant{
				Grantee:    &s3types.Grantee{Type: s3types.TypeCanonicalUser, ID: curACL.Owner.ID},
				Permission: s3types.PermissionFullControl,
			})
		}
	}

	if _, err := c.S3Client.PutBucketAcl(ctx, &s3ACL); err != nil {
		return err
	}

	return nil
}

func (c *Client) SetObjectACL(ctx context.Context, bucket, key string, acl *ACL) error {
	s3ACL := s3.PutObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if acl.Canned != "" {
		if !utils.IsInList(ObjectCannedACLToStrings(), acl.Canned) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl.Canned,
				strings.Join(ObjectCannedACLToStrings(), ", "))
		}

		s3ACL.ACL = s3types.ObjectCannedACL(acl.Canned)
	} else {
		s3ACL.AccessControlPolicy = &s3types.AccessControlPolicy{Grants: acl.toS3Grants()}

		// As a safety precaution, if the caller didn't explicitly set a Grantee
		// with the FULL_CONTROL permission we set it to the current object owner.
		if acl.FullControl == "" {
			curACL, err := c.S3Client.GetObjectAcl(ctx, &s3.GetObjectAclInput{
				Bucket: s3ACL.Bucket,
				Key:    s3ACL.Key,
			})
			if err != nil {
				return fmt.Errorf("unable to retrieve current ACL: %w", err)
			}

			s3ACL.AccessControlPolicy.Grants = append(s3ACL.AccessControlPolicy.Grants, s3types.Grant{
				Grantee:    &s3types.Grantee{Type: s3types.TypeCanonicalUser, ID: curACL.Owner.ID},
				Permission: s3types.PermissionFullControl,
			})
		}
	}

	if _, err := c.S3Client.PutObjectAcl(ctx, &s3ACL); err != nil {
		return err
	}

	return nil
}

func (c *Client) SetObjectsACL(ctx context.Context, bucket, prefix string, acl *ACL, recursive bool) error {
	return c.ForEachObjectUnfiltered(ctx, bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.SetObjectACL(ctx, bucket, aws.ToString(o.Key), acl)
	})
}

func BucketCannedACLToStrings() []string {
	s3BucketCannedACLs := s3types.BucketCannedACLPrivate.Values()

	list := make([]string, len(s3BucketCannedACLs))
	for i, v := range s3BucketCannedACLs {
		list[i] = string(v)
	}

	return list
}

func ObjectCannedACLToStrings() []string {
	s3ObjectCannedACLs := s3types.ObjectCannedACLPrivate.Values()

	list := make([]string, len(s3ObjectCannedACLs))
	for i, v := range s3ObjectCannedACLs {
		list[i] = string(v)
	}

	return list
}

// toS3Grants converts the local ACL representation to the s3types.Grant format.
func (a ACL) toS3Grants() []s3types.Grant {
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

// ACLFromS3 converts an S3 ACL Grant to the local ACL representation.
func ACLFromS3(grants []s3types.Grant) ACL {
	acl := ACL{
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
	case ACLGranteeAllUsers:
		return "ALL_USERS"

	case ACLGranteeAuthenticatedUsers:
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
			URI:  aws.String(ACLGranteeAllUsers),
		}

	case "AUTHENTICATED_USERS":
		return &s3types.Grantee{
			Type: s3types.TypeGroup,
			URI:  aws.String(ACLGranteeAuthenticatedUsers),
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
		case ACLGranteeAllUsers:
			return aws.String("uri=" + ACLGranteeAllUsers)

		case ACLGranteeAuthenticatedUsers:
			return aws.String("uri=" + ACLGranteeAuthenticatedUsers)
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
