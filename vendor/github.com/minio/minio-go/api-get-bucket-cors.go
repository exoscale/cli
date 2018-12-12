/*
 * Minio Go Library for Amazon S3 Compatible Cloud Storage
 * Copyright 2015-2018 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package minio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/minio/minio-go/pkg/s3utils"
)

type corsConfiguration struct {
	CORSRule []CORSRule `xml:"CORSRule"`
}

// GetBucketCORS fetches the Bucket CORS metadata
func (c Client) GetBucketCORS(bucketName string) (*BucketInfo, error) {
	return c.getBucketCORSWithContext(context.Background(), bucketName)
}

// getBucketCORSWithContext
func (c Client) getBucketCORSWithContext(ctx context.Context, bucketName string) (*BucketInfo, error) {
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return nil, err
	}

	found, err := c.BucketExists(bucketName)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("bucket %q not found", bucketName)
	}

	query := url.Values{}
	query.Add("cors", "")
	resp, err := c.executeMethod(ctx, "GET", requestMetadata{
		bucketName:  bucketName,
		queryValues: query,
	})
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, httpRespToErrorResponse(resp, bucketName, "")
	}

	cors := &corsConfiguration{}
	if err := xmlDecoder(resp.Body, cors); err != nil {
		return nil, err
	}

	return &BucketInfo{
		Name: bucketName,
		CORS: cors.CORSRule,
	}, nil
}
