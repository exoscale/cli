package sos

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/exoscale/cli/utils"
)

func (c *Client) setBucketACL(bucket string, acl *storageACL) error {
	s3ACL := s3.PutBucketAclInput{Bucket: aws.String(bucket)}

	if acl.Canned != "" {
		if !utils.IsInList(s3BucketCannedACLToStrings(), acl.Canned) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl.Canned,
				strings.Join(s3BucketCannedACLToStrings(), ", "))
		}

		s3ACL.ACL = s3types.BucketCannedACL(acl.Canned)
	} else {
		s3ACL.AccessControlPolicy = &s3types.AccessControlPolicy{Grants: acl.toS3Grants()}

		// As a safety precaution, if the caller didn't explicitly set a Grantee
		// with the FULL_CONTROL permission we set it to the current bucket owner.
		if acl.FullControl == "" {
			curACL, err := c.GetBucketAcl(gContext, &s3.GetBucketAclInput{Bucket: aws.String(bucket)})
			if err != nil {
				return fmt.Errorf("unable to retrieve current ACL: %w", err)
			}

			s3ACL.AccessControlPolicy.Grants = append(s3ACL.AccessControlPolicy.Grants, s3types.Grant{
				Grantee:    &s3types.Grantee{Type: s3types.TypeCanonicalUser, ID: curACL.Owner.ID},
				Permission: s3types.PermissionFullControl,
			})
		}
	}

	if _, err := c.PutBucketAcl(gContext, &s3ACL); err != nil {
		return err
	}

	return nil
}

func (c *Client) setObjectACL(bucket, key string, acl *storageACL) error {
	s3ACL := s3.PutObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if acl.Canned != "" {
		if !utils.IsInList(s3ObjectCannedACLToStrings(), acl.Canned) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl.Canned,
				strings.Join(s3ObjectCannedACLToStrings(), ", "))
		}

		s3ACL.ACL = s3types.ObjectCannedACL(acl.Canned)
	} else {
		s3ACL.AccessControlPolicy = &s3types.AccessControlPolicy{Grants: acl.toS3Grants()}

		// As a safety precaution, if the caller didn't explicitly set a Grantee
		// with the FULL_CONTROL permission we set it to the current object owner.
		if acl.FullControl == "" {
			curACL, err := c.GetObjectAcl(gContext, &s3.GetObjectAclInput{
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

	if _, err := c.PutObjectAcl(gContext, &s3ACL); err != nil {
		return err
	}

	return nil
}

func (c *Client) setObjectsACL(bucket, prefix string, acl *storageACL, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.setObjectACL(bucket, aws.ToString(o.Key), acl)
	})
}
