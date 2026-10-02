package sos_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/storage/sos"
)

// fakeSOSServer returns an httptest.Server that responds to any HeadBucket
// request with a 200 and the given zone in X-Amz-Bucket-Region.
// The returned channel receives every Authorization header observed.
func fakeSOSServer(t *testing.T, zone string) (*httptest.Server, <-chan string) {
	t.Helper()
	authHeaders := make(chan string, 8)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeaders <- r.Header.Get("Authorization")
		w.Header().Set("X-Amz-Bucket-Region", zone)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, authHeaders
}

// commonOptFns returns the minimal CommonConfigOptFns used in production
// (User-Agent only, no credentials - same as cmd/storage/storage.go).
func commonOptFns() []func(*awsconfig.LoadOptions) error {
	return []func(*awsconfig.LoadOptions) error{
		awsconfig.WithAPIOptions([]func(*middleware.Stack) error{
			awsmiddleware.AddUserAgentKeyValue("Exoscale-CLI", "test"),
		}),
	}
}

// TestClientOptZoneFromBucket_CredentialsSigned verifies that after the fix,
// ClientOptZoneFromBucket sends a signed (authenticated) HeadBucket request
// rather than falling through to the EC2 IMDS credential chain.
func TestClientOptZoneFromBucket_CredentialsSigned(t *testing.T) {
	const (
		testKey    = "EXOtestkey"
		testSecret = "testsecret"
		testZone   = "ch-gva-2"
		testBucket = "my-bucket"
	)

	srv, authHeaders := fakeSOSServer(t, testZone)

	account.CurrentAccount = &account.Account{
		Key:         testKey,
		Secret:      testSecret,
		DefaultZone: testZone,
		SosEndpoint: srv.URL, // no {zone} placeholder needed for the discovery request
	}
	sos.CommonConfigOptFns = commonOptFns()
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	_, err := sos.NewStorageClient(
		context.Background(),
		sos.ClientOptZoneFromBucket(context.Background(), testBucket),
	)
	require.NoError(t, err)

	authHeader := <-authHeaders
	assert.NotEmpty(t, authHeader, "HeadBucket should carry an Authorization header")
	assert.True(t, strings.Contains(authHeader, testKey),
		"Authorization header should reference the configured API key, got: %s", authHeader)
}

// TestClientOptZoneFromBucket_ZoneDiscovered verifies that the zone returned
// by the fake SOS server is correctly assigned to the client.
func TestClientOptZoneFromBucket_ZoneDiscovered(t *testing.T) {
	const (
		testZone   = "ch-dk-2"
		testBucket = "my-bucket"
	)

	srv, _ := fakeSOSServer(t, testZone)

	account.CurrentAccount = &account.Account{
		Key:         "EXOtest",
		Secret:      "secret",
		DefaultZone: "ch-gva-2", // intentionally different from testZone
		SosEndpoint: srv.URL,
	}
	sos.CommonConfigOptFns = commonOptFns()
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	client, err := sos.NewStorageClient(
		context.Background(),
		sos.ClientOptZoneFromBucket(context.Background(), testBucket),
	)
	require.NoError(t, err)
	assert.Equal(t, testZone, client.Zone,
		"client Zone should be set to the zone returned by the SOS endpoint")
}

// TestClientOptZoneFromBucket_NoIMDSFallback is a regression test for the
// bug introduced by the aws-sdk-go-v2 v1.2.0 -> v1.23.1 bump: in the new SDK
// GetBucketRegion no longer forces AnonymousCredentials, so without explicit
// credentials the default chain falls through to EC2 IMDS and fails on
// non-EC2 hosts.  This test asserts the fixed code never produces an
// IMDS-related error.
func TestClientOptZoneFromBucket_NoIMDSFallback(t *testing.T) {
	srv, _ := fakeSOSServer(t, "ch-gva-2")

	account.CurrentAccount = &account.Account{
		Key:         "EXOtest",
		Secret:      "secret",
		DefaultZone: "ch-gva-2",
		SosEndpoint: srv.URL,
	}
	sos.CommonConfigOptFns = commonOptFns()

	// Disable IMDS to make any fallback immediately fatal rather than slow.
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	// Also clear any AWS env-var credentials so the only source is the static
	// provider we expect ClientOptZoneFromBucket to wire in.
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/dev/null")

	_, err := sos.NewStorageClient(
		context.Background(),
		sos.ClientOptZoneFromBucket(context.Background(), "my-bucket"),
	)
	require.NoError(t, err, "must not fall through to EC2 IMDS when static credentials are available")
}

// TestCopyObject_CopySourceEncoding verifies that the copy source of an
// in-place copy (headers/metadata manipulation) is URL-encoded: the server
// decodes it, so a raw key containing "+" or a percent-encoded sequence would
// point at another, non-existent object.
func TestCopyObject_CopySourceEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "plain key",
			key:      "test-key",
			expected: "test-bucket/test-key",
		},
		{
			name:     "ampersand",
			key:      "a-&-b.png",
			expected: "test-bucket/a-%26-b.png",
		},
		{
			name:     "literal percent-encoded sequence",
			key:      "a-%26-b.png",
			expected: "test-bucket/a-%2526-b.png",
		},
		{
			name:     "space",
			key:      "a b.png",
			expected: "test-bucket/a+b.png",
		},
		{
			name:     "plus",
			key:      "a+b.png",
			expected: "test-bucket/a%2Bb.png",
		},
		{
			name:     "all of them in a nested key",
			key:      "dir/a-&-b.png/c d+e-%26-f.webp",
			expected: "test-bucket/dir%2Fa-%26-b.png%2Fc+d%2Be-%2526-f.webp",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &sos.Client{
				S3Client: &MockS3API{
					mockGetObject: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{}, nil
					},
					mockGetObjectAcl: func(ctx context.Context, input *s3.GetObjectAclInput, optFns ...func(*s3.Options)) (*s3.GetObjectAclOutput, error) {
						return &s3.GetObjectAclOutput{
							Owner: &types.Owner{
								ID: aws.String("sarah"),
							},
						}, nil
					},
				},
			}

			object, err := client.CopyObject(context.Background(), "test-bucket", tc.key)
			require.NoError(t, err)
			assert.Equal(t, tc.key, *object.Key)
			assert.Equal(t, tc.expected, *object.CopySource)

			// Decoding the copy source the way the server does must give
			// the key back.
			decoded, err := url.QueryUnescape(*object.CopySource)
			require.NoError(t, err)
			assert.Equal(t, "test-bucket/"+tc.key, decoded)
		})
	}
}
