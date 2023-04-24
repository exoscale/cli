package sos

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	// storageCommonConfigOptFns represents the list of AWS SDK configuration options common
	// to all commands. In addition to those, some commands can/must set additional options
	// specific to their execution context.
	storageCommonConfigOptFns []func(*awsconfig.LoadOptions) error
)

type Client struct {
	*s3.Client

	zone string
}

// forEachObject is a convenience wrapper to execute a callback function on
// each object listed in the specified bucket/prefix. Upon callback function
// error, the whole processing ends.
func (c *Client) ForEachObject(ctx context.Context, bucket, prefix string, recursive bool, fn func(*s3types.Object) error) error {
	// The "/" value can be used at command-level to mean that we want to
	// list from the root of the bucket, but the actual bucket root is an
	// empty prefix.
	if prefix == "/" {
		prefix = ""
	}

	dirs := make(map[string]struct{})

	var ct string
	for {
		res, err := c.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: aws.String(ct),
		})
		if err != nil {
			return err
		}
		ct = aws.ToString(res.NextContinuationToken)

		for _, o := range res.Contents {
			// If not invoked in recursive mode, split object keys on the "/" separator and skip
			// objects "below" the base directory prefix.
			parts := strings.SplitN(strings.TrimPrefix(aws.ToString(o.Key), prefix), "/", 2)
			if len(parts) > 1 && !recursive {
				dir := path.Base(parts[0])
				if _, ok := dirs[dir]; !ok {
					dirs[dir] = struct{}{}
				}
				continue
			}

			// If the prefix doesn't end with a trailing prefix separator ("/"),
			// consider it as a single object key and match only one exact result
			// (except in recursive mode, where the prefix is expected to be a
			// "directory").
			if !recursive && !strings.HasSuffix(prefix, "/") && aws.ToString(o.Key) != prefix {
				continue
			}

			o := o
			if err := fn(&o); err != nil {
				return err
			}
		}

		if !res.IsTruncated {
			break
		}
	}

	return nil
}

// copyObject is a helper function to be used in commands involving object
// copying such as metadata/headers manipulation, retrieving information about
// the targeted object for a later copy.
func (c *Client) CopyObject(ctx context.Context, bucket, key string) (*s3.CopyObjectInput, error) {
	srcObject, err := c.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object information: %w", err)
	}

	// Object ACL are reset during a CopyObject operation,
	// we must set them explicitly on the copied object.

	acl, err := c.GetObjectAcl(ctx, &s3.GetObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object ACL: %w", err)
	}

	copyObject := s3.CopyObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(key),
		CopySource:        aws.String(bucket + "/" + key),
		Metadata:          srcObject.Metadata,
		MetadataDirective: s3types.MetadataDirectiveReplace,

		// Headers
		CacheControl:       srcObject.CacheControl,
		ContentDisposition: srcObject.ContentDisposition,
		ContentEncoding:    srcObject.ContentEncoding,
		ContentLanguage:    srcObject.ContentLanguage,
		ContentType:        srcObject.ContentType,
		Expires:            srcObject.Expires,
	}

	storageACLToCopyObject(acl, &copyObject)

	return &copyObject, nil
}

type ClientOpt func(*Client) error

func ClientOptWithZone(zone string) ClientOpt {
	return func(c *Client) error { c.zone = zone; return nil }
}

func ClientOptZoneFromBucket(ctx context.Context, zone, endpoint, bucket string) ClientOpt {
	return func(c *Client) error {
		cfg, err := awsconfig.LoadDefaultConfig(
			ctx,
			append(storageCommonConfigOptFns,
				awsconfig.WithEndpointResolver(aws.EndpointResolverFunc(
					func(service, region string) (aws.Endpoint, error) {
						sosURL := strings.Replace(
							endpoint,
							"{zone}",
							zone,
							1,
						)
						return aws.Endpoint{URL: sosURL}, nil
					})),
			)...)
		if err != nil {
			return err
		}

		region, err := s3manager.GetBucketRegion(ctx, s3.NewFromConfig(cfg), bucket, func(o *s3.Options) {
			o.UsePathStyle = true
		})
		if err != nil {
			return err
		}

		c.zone = region
		return nil
	}
}

type Account struct {
	Name                 string
	Account              string
	Endpoint             string
	ComputeEndpoint      string // legacy config.
	DNSEndpoint          string
	SosEndpoint          string
	RunstatusEndpoint    string
	Environment          string
	Key                  string
	Secret               string
	SecretCommand        []string
	DefaultZone          string
	DefaultSSHKey        string
	DefaultTemplate      string
	DefaultRunstatusPage string
	DefaultOutputFormat  string
	ClientTimeout        int
	CustomHeaders        map[string]string
}

func NewStorageClient(ctx context.Context, account *Account, opts ...ClientOpt) (*Client, error) {
	var (
		client = Client{
			zone: account.DefaultZone,
		}

		caCerts io.Reader
	)

	for _, opt := range opts {
		if err := opt(&client); err != nil {
			return nil, fmt.Errorf("option error: %w", err)
		}
	}

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		append(storageCommonConfigOptFns,
			awsconfig.WithRegion(client.zone),

			awsconfig.WithEndpointResolver(aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					sosURL := strings.Replace(account.SosEndpoint, "{zone}", client.zone, 1)
					return aws.Endpoint{
						URL:           sosURL,
						SigningRegion: client.zone,
					}, nil
				})),

			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				account.Key,
				account.APISecret(),
				"")),

			awsconfig.WithCustomCABundle(caCerts),
		)...)
	if err != nil {
		return nil, err
	}

	client.Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &client, nil
}
