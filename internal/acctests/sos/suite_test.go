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

	testBucketName := fmt.Sprintf("exo-cli-acc-tests-%d", rand.Int())
	s.BucketName = testBucketName
	input := &s3.CreateBucketInput{
		Bucket: &s.BucketName,
	}

	s.S3Client = s3.NewFromConfig(cfg)

	_, err = s.S3Client.CreateBucket(ctx, input)
	s.Assert().NoError(err, fmt.Sprintf("error creating test bucket %q", s.BucketName))
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

func getStr(a *string) string {
	if a == nil {
		return ""
	}

	return *a
}

// TODO(sauterp) once deletion of versioned buckets, tests should handle deletions themselves.
func (s *SOSSuite) deleteBucket(bucketName string) {
	ctx := context.Background()

	// Delete all object versions
	listObjectsVersionsInput := &s3.ListObjectVersionsInput{
		Bucket: &bucketName,
	}

	listObjectsVersionsResp, err := s.S3Client.ListObjectVersions(ctx, listObjectsVersionsInput)
	s.Assert().NoError(err, fmt.Sprintf("error deleting test bucket %q", bucketName))

	for _, obj := range listObjectsVersionsResp.Versions {
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket:    &bucketName,
			Key:       obj.Key,
			VersionId: obj.VersionId,
		}
		_, err := s.S3Client.DeleteObject(ctx, deleteObjectInput)
		s.Assert().NoError(err, fmt.Sprintf("deleting object %s version %s", getStr(obj.Key), getStr(obj.VersionId)))
	}

	// Delete bucket
	deleteBucketInput := &s3.DeleteBucketInput{
		Bucket: &bucketName,
	}

	_, err = s.S3Client.DeleteBucket(ctx, deleteBucketInput)
	s.Assert().NoError(err)
}

func TestSOSSuite(t *testing.T) {
	s := &SOSSuite{
		ExoCLIExecutable: "../../../bin/exo",
	}

	suite.Run(t, s)
}

func (s *SOSSuite) writeFile(filename, content string) {
	err := os.WriteFile(s.PrepDir+filename, []byte(content), 0644)
	s.Assert().NoError(err)
}

func (s *SOSSuite) exo(cmdStr string) (string, error) {
	// remove the "exo " prefix
	cmdStr = cmdStr[4:]

	cmdWithBucket := strings.Replace(cmdStr, "{bucket}", s.BucketName, 1)
	cmdWithPrepDir := strings.Replace(cmdWithBucket, "{prepDir}", s.PrepDir, 1)
	cmdComplete := strings.Replace(cmdWithPrepDir, "{downloadDir}", s.DownloadDir, 1)
	cmds := strings.Split(cmdComplete, " ")

	cmds = append([]string{"--quiet"}, cmds...)

	s.T().Logf("executing command: exo %s", strings.Join(cmds, " "))

	command := exec.Command(s.ExoCLIExecutable, cmds...)
	output, err := command.CombinedOutput()
	if len(output) > 0 {
		s.T().Log(string(output))
	}

	return string(output), err
}
