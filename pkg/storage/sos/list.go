package sos

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/exoscale/cli/pkg/storage/sos/object"
)

type listFunc[ObjectType object.ObjectInterface] func(ctx context.Context) (*listCallOut[ObjectType], error)

type listCallOut[ObjectType object.ObjectInterface] struct {
	Objects        []ObjectType
	CommonPrefixes []string
	IsTruncated    bool
}

type ObjectListing[ObjectType object.ObjectInterface] struct {
	List           []ObjectType
	CommonPrefixes []string
}

func (c *Client) ListObjectsFunc(bucket, prefix string, recursive, stream bool) listFunc[object.ObjectInterface] {
	var continuationToken *string

	deduplicate := GetCommonPrefixDeduplicator(stream)

	return func(ctx context.Context) (*listCallOut[object.ObjectInterface], error) {
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
			objects = append(objects, o)
		}

		return &listCallOut[object.ObjectInterface]{
			Objects:        objects,
			CommonPrefixes: deduplicate(res.CommonPrefixes),
			IsTruncated:    res.IsTruncated,
		}, nil
	}
}

func (c *Client) ListVersionedObjectsFunc(bucket, prefix string, recursive, stream bool) listFunc[object.ObjectVersionInterface] {
	var keyMarker *string
	var versionIdMarker *string

	deduplicate := GetCommonPrefixDeduplicator(stream)

	return func(ctx context.Context) (*listCallOut[object.ObjectVersionInterface], error) {
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

		var objects []object.ObjectVersionInterface
		for i := range res.Versions {
			o := object.ObjectVersion{
				ObjectVersion: &res.Versions[i],
			}

			objects = append(objects, &o)
		}

		return &listCallOut[object.ObjectVersionInterface]{
			Objects:        objects,
			CommonPrefixes: deduplicate(res.CommonPrefixes),
			IsTruncated:    res.IsTruncated,
		}, nil
	}
}

func assignVersionNumbers(objs []object.ObjectVersionInterface) {
	// S3 does not guarantee that versions of objects appear in a particular order thus we have to sort before we assign a version number
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].GetLastModified().After(*objs[j].GetLastModified())
	})

	latestVersionPerObj := make(map[string]uint64)
	// we traverse in reverse order because the latest version should always get the highest version number. Why don't we sort in reverse order? Because we also want the latest version to appear on top.
	for i := len(objs) - 1; i >= 0; i-- {
		obj := objs[i]
		key := *obj.GetKey()
		latestVersion, ok := latestVersionPerObj[key]
		if !ok {
			latestVersionPerObj[key] = 0
			obj.SetVersionNumber(0)

			continue
		}

		latestVersion++
		latestVersionPerObj[key] = latestVersion
		obj.SetVersionNumber(latestVersion)
	}
}

func getObjectListing[ObjectType object.ObjectInterface](ctx context.Context, c *Client, list listFunc[ObjectType], stream bool) (*ObjectListing[ObjectType], error) {
	listing := ObjectListing[ObjectType]{}

	for {
		res, err := list(ctx)
		if err != nil {
			return nil, err
		}

		if stream {
			for _, o := range res.Objects {
				fmt.Println(*o.GetKey())
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

func (c *Client) ListObjects(ctx context.Context, list listFunc[object.ObjectInterface], recursive, stream bool, filters []object.ObjectFilterFunc) (*object.ListObjectsOutput, error) {
	listing, err := getObjectListing(ctx, c, list, stream)
	if err != nil {
		return nil, err
	}

	var objects []object.ObjectInterface
	for _, obj := range listing.List {
		if object.ApplyFilters(obj, filters) {
			objects = append(objects, obj)
		}
	}

	listing.List = objects

	return prepareListObjectsOutput(listing, recursive, stream)
}

func (c *Client) ListObjectsVersions(ctx context.Context, list listFunc[object.ObjectVersionInterface], recursive, stream bool,
	filters []object.ObjectFilterFunc,
	versionFilters []object.ObjectVersionFilterFunc) (*object.ListObjectsOutput, error) {
	listing, err := getObjectListing(ctx, c, list, stream)
	if err != nil {
		return nil, err
	}

	assignVersionNumbers(listing.List)

	var objects []object.ObjectVersionInterface
	for _, obj := range listing.List {
		if object.ApplyFilters(obj, filters) && object.ApplyVersionedFilters(obj, versionFilters) {
			objects = append(objects, obj)
		}
	}

	listing.List = objects

	return prepareListObjectsOutput(listing, recursive, stream)
}

func prepareListObjectsOutput[ObjectType object.ObjectInterface](listing *ObjectListing[ObjectType], recursive, stream bool) (*object.ListObjectsOutput, error) {
	out := make(object.ListObjectsOutput, 0)
	dirsOut := make(object.ListObjectsOutput, 0) // to separate common prefixes (folders) from objects (files)

	if !recursive {
		for _, cp := range listing.CommonPrefixes {
			dirsOut = append(dirsOut, object.ListObjectsItemOutput{
				Path: cp,
				Dir:  true,
			})
		}
	}

	for _, o := range listing.List {
		out = append(out, *o.GetListObjectsItemOutput())
	}

	// To be user friendly, we are going to push dir records to the top of the output list
	if !stream && !recursive {
		out = append(dirsOut, out...)
	}

	return &out, nil
}
