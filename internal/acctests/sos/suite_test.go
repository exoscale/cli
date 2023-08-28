package sos_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"

	"github.com/exoscale/cli/internal/acctests"
)

const (
	zone = "ch-dk-2"
)

// SOSSuite creates a bucket with a partially random name
// as well as two temporary directories for preparations and downloads respectively,
// for each test in the suite.
type SOSSuite struct {
	acctests.AcceptanceTestSuite

	BucketName       string
	ExoCLIExecutable string

	PrepDir     string
	DownloadDir string

	ObjectList []string

	S3Client *s3.Client
}

func (s *SOSSuite) SetupTest() {
	ctx := context.Background()

	var err error

	tmpDirPrefix := "exo-cli-acc-tests"
	prepDir, err := ioutil.TempDir("", tmpDirPrefix)
	s.Assert().NoError(err)
	s.PrepDir = prepDir + "/"

	downloadDir, err := ioutil.TempDir("", tmpDirPrefix)
	s.Assert().NoError(err)
	s.DownloadDir = downloadDir + "/"

	var caCerts io.Reader

	cfg, err := config.LoadDefaultConfig(
		ctx,
		append([]func(*config.LoadOptions) error{},
			config.WithRegion(zone),

			config.WithEndpointResolver(aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					sosURL := fmt.Sprintf("https://sos-%s.exo.io", zone)
					return aws.Endpoint{
						URL:           sosURL,
						SigningRegion: zone,
					}, nil
				})),

			config.WithCustomCABundle(caCerts),
		)...)
	s.Assert().NoError(err)

	input := &s3.CreateBucketInput{
		Bucket: &s.BucketName,
	}

	s.S3Client = s3.NewFromConfig(cfg)

	s.T().Logf("creating test bucket %q", s.BucketName)
	_, err = s.S3Client.CreateBucket(ctx, input)
	s.Assert().NoError(err)
}

func (s *SOSSuite) TearDownTest() {
	var err error

	s.deleteBucket(s.BucketName)
	s.Assert().NoError(err)

	err = os.RemoveAll(s.PrepDir)
	s.Assert().NoError(err)

	err = os.RemoveAll(s.DownloadDir)
	s.Assert().NoError(err)
}

func (s *SOSSuite) deleteBucket(bucketName string) {
	ctx := context.Background()

	s.T().Logf("deleting test bucket %q", bucketName)

	// Delete all objects
	listObjectsInput := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}

	listObjectsResp, err := s.S3Client.ListObjectsV2(ctx, listObjectsInput)
	s.Assert().NoError(err)

	for _, obj := range listObjectsResp.Contents {
		s.T().Log("deleting object: ", *obj.Key)
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: &bucketName,
			Key:    obj.Key,
		}
		_, err := s.S3Client.DeleteObject(ctx, deleteObjectInput)
		s.Assert().NoError(err)
	}

	// Delete all object versions
	listObjectsVersionsInput := &s3.ListObjectVersionsInput{
		Bucket: &bucketName,
	}

	listObjectsVersionsResp, err := s.S3Client.ListObjectVersions(ctx, listObjectsVersionsInput)
	s.Assert().NoError(err)

	for _, obj := range listObjectsVersionsResp.Versions {
		s.T().Log("deleting object version: ", *obj.Key, *obj.VersionId)
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket:    &bucketName,
			Key:       obj.Key,
			VersionId: obj.VersionId,
		}
		_, err := s.S3Client.DeleteObject(ctx, deleteObjectInput)
		s.Assert().NoError(err)
	}

	// Delete bucket
	deleteBucketInput := &s3.DeleteBucketInput{
		Bucket: &bucketName,
	}

	_, err = s.S3Client.DeleteBucket(ctx, deleteBucketInput)
	s.Assert().NoError(err)
}

func TestSOSSuite(t *testing.T) {
	testBucketName := fmt.Sprintf("exo-cli-acc-tests-%d", rand.Int())

	s := &SOSSuite{
		BucketName:       testBucketName,
		ExoCLIExecutable: "../../../bin/exo",
	}
	suite.Run(t, s)
}

func (s *SOSSuite) writeFile(filename, content string) {
	err := os.WriteFile(s.PrepDir+filename, []byte(content), 0644)
	s.Assert().NoError(err)
}

func (s *SOSSuite) exo(cmdStr string) string {
	cmdWithBucket := strings.Replace(cmdStr, "{bucket}", s.BucketName, 1)
	cmdWithPrepDir := strings.Replace(cmdWithBucket, "{prepDir}", s.PrepDir, 1)
	cmdComplete := strings.Replace(cmdWithPrepDir, "{downloadDir}", s.DownloadDir, 1)
	cmds := strings.Split(cmdComplete, " ")

	s.T().Logf("executing command: exo %s", strings.Join(cmds, " "))
	command := exec.Command(s.ExoCLIExecutable, cmds...)
	output, err := command.CombinedOutput()
	s.T().Log(string(output))
	s.Assert().NoError(err)

	return string(output)
}
