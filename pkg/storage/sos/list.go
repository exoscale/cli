package sos

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/exoscale/cli/pkg/storage/sos/object"
)

type listFunc func(ctx context.Context) (*listCallOut, error)

type listCallOut struct {
	Objects        []object.ObjectInterface
	CommonPrefixes []string
	IsTruncated    bool
}

type ObjectListing struct {
	List           []object.ObjectInterface
	CommonPrefixes []string
}

func (c *Client) ListObjectsFunc(bucket, prefix string, recursive, stream bool, filters []object.ObjectFilterFunc) listFunc {
	var continuationToken *string

	deduplicate := GetCommonPrefixDeduplicator(stream)

	return func(ctx context.Context) (*listCallOut, error) {
		req := s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		if !recursive {
			req.Delimiter = aws.String("/")
		}

		res, err := c.S3Client.ListObjectsV2(ctx, &req)
		if err != nil {
			return nil, err
		}

		continuationToken = res.NextContinuationToken

		var objects []object.ObjectInterface
		for i := range res.Contents {
			o := &object.Object{
				Object: &res.Contents[i],
			}

			if object.ApplyFilters(o, filters) {
				objects = append(objects, o)
			}
		}

		return &listCallOut{
			Objects:        objects,
			CommonPrefixes: deduplicate(res.CommonPrefixes),
			IsTruncated:    res.IsTruncated,
		}, nil
	}
}

func (c *Client) ListVersionedObjectsFunc(bucket, prefix string, recursive, stream bool,
	filters []object.ObjectFilterFunc,
	versionFilters []object.ObjectVersionFilterFunc) listFunc {
	var keyMarker *string
	var versionIdMarker *string

	deduplicate := GetCommonPrefixDeduplicator(stream)

	return func(ctx context.Context) (*listCallOut, error) {
		req := s3.ListObjectVersionsInput{
			Bucket:          aws.String(bucket),
			Prefix:          aws.String(prefix),
			KeyMarker:       keyMarker,
			VersionIdMarker: versionIdMarker,
		}

		if !recursive {
			req.Delimiter = aws.String("/")
		}

		res, err := c.S3Client.ListObjectVersions(ctx, &req)
		if err != nil {
			return nil, err
		}

		keyMarker = res.NextKeyMarker
		versionIdMarker = res.NextVersionIdMarker

		var objects []object.ObjectInterface
		for i := range res.Versions {
			o := object.ObjectVersion{
				ObjectVersion: &res.Versions[i],
			}

			if object.ApplyFilters(&o, filters) && object.ApplyVersionedFilters(&o, versionFilters) {
				objects = append(objects, &o)
			}
		}

		return &listCallOut{
			Objects:        objects,
			CommonPrefixes: deduplicate(res.CommonPrefixes),
			IsTruncated:    res.IsTruncated,
		}, nil
	}
}

func (c *Client) GetObjectListing(ctx context.Context, list listFunc, stream bool) (*ObjectListing, error) {
	listing := ObjectListing{}

	for {
		res, err := list(ctx)
		if err != nil {
			return nil, err
		}

		if stream {
			for _, o := range res.Objects {
				fmt.Println(o.GetKey())
			}
		} else {
			listing.List = append(listing.List, res.Objects...)
		}

		listing.CommonPrefixes = append(listing.CommonPrefixes, res.CommonPrefixes...)

		if !res.IsTruncated {
			break
		}
	}

	return &listing, nil
}

func (c *Client) ListObjects(ctx context.Context, list listFunc, recursive, stream bool) (*ListObjectsOutput, error) {
	listing, err := c.GetObjectListing(ctx, list, stream)
	if err != nil {
		return nil, err
	}

	return c.prepareListObjectsOutput(listing, recursive, stream)
}

func (c *Client) prepareListObjectsOutput(listing *ObjectListing, recursive, stream bool) (*ListObjectsOutput, error) {
	out := make(ListObjectsOutput, 0)
	dirsOut := make(ListObjectsOutput, 0) // to separate common prefixes (folders) from objects (files)

	if !recursive {
		for _, cp := range listing.CommonPrefixes {
			dirsOut = append(dirsOut, ListObjectsItemOutput{
				Path: cp,
				Dir:  true,
			})
		}
	}

	for _, o := range listing.List {
		out = append(out, ListObjectsItemOutput{
			Path:         aws.ToString(o.GetKey()),
			Size:         o.GetSize(),
			LastModified: o.GetLastModified().Format(TimestampFormat),
		})
	}

	// To be user friendly, we are going to push dir records to the top of the output list
	if !stream && !recursive {
		out = append(dirsOut, out...)
	}

	return &out, nil
}
