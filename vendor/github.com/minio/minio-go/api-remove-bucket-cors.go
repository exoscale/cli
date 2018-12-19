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

// RemoveBucketCORS fetches the Bucket CORS metadata
func (c Client) RemoveBucketCORS(bucketName string) error {
	return c.removeBucketCORSWithContext(context.Background(), bucketName)
}

// removeBucketCORSWithContext
func (c Client) removeBucketCORSWithContext(ctx context.Context, bucketName string) error {
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return err
	}

	found, err := c.BucketExists(bucketName)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("bucket %q not found", bucketName)
	}

	query := url.Values{}
	query.Add("cors", "")
	resp, err := c.executeMethod(ctx, "DELETE", requestMetadata{
		bucketName:  bucketName,
		queryValues: query,
	})
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp != nil && resp.StatusCode != http.StatusNoContent {
		return httpRespToErrorResponse(resp, bucketName, "")
	}

	return nil
}
