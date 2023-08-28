package main

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
)

const (
	zone = "ch-dk-2"
)

type SOSSuite struct {
	suite.Suite

	BucketName       string
	ExoCLIExecutable string

	PrepDir     string
	DownloadDir string

	ObjectList []string

	S3Client *s3.Client
}

func (s *SOSSuite) SetupTest() {
	// TODO introduce a TF_ACC like test guard
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
	_, err = s.S3Client.CreateBucket(ctx, input)
	s.Assert().NoError(err)
}

func (s *SOSSuite) TearDownTest() {
	var (
		err error
		ctx context.Context = context.Background()
	)

	for _, v := range s.ObjectList {
		s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &s.BucketName,
			Key:    &v,
		})
	}

	_, err = s.S3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: &s.BucketName,
	})
	s.Assert().NoError(err)

	err = os.RemoveAll(s.PrepDir)
	s.Assert().NoError(err)

	err = os.RemoveAll(s.DownloadDir)
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

func (s *SOSSuite) TestDownloadFiles() {
	file1Name := "file1.txt"
	originalContent := "original content"

	s.writeFile(file1Name, originalContent)
	s.uploadFile(file1Name)

	expectedContent := "expected new content"
	s.writeFile(file1Name, expectedContent)
	s.uploadFile(file1Name)

	s.downloadVersion(s.findLatestVersion())

	files, err := os.ReadDir(s.DownloadDir)
	s.Assert().NoError(err)
	s.Equal(1, len(files))
	s.Equal(file1Name, files[0].Name())

	file1Content, err := os.ReadFile(s.DownloadDir + file1Name)
	s.Assert().NoError(err)
	s.Equal(expectedContent, string(file1Content))
}

func (s *SOSSuite) findLatestVersion() string {
	output := s.exo("storage list --versions " + s.BucketName)

	lines := strings.Split(output, "\n")
	words := strings.Split(lines[0], " ")
	return words[len(words)-1]
}

func (s *SOSSuite) downloadVersion(version string) {
	s.exo(fmt.Sprintf("storage download --only-versions %s -r %s %s", version, s.BucketName, s.DownloadDir))
}

func (s *SOSSuite) writeFile(filename, content string) {
	err := os.WriteFile(s.PrepDir+filename, []byte(content), 0644)
	s.Assert().NoError(err)
}

func (s *SOSSuite) uploadFile(filePath string) {
	s.exo(fmt.Sprintf("storage upload %s %s", s.PrepDir+filePath, s.BucketName))
	s.ObjectList = append(s.ObjectList, filePath)
}

func (s *SOSSuite) exo(args string) string {
	cmds := strings.Split(args, " ")

	cmds = append([]string{"-A", "exosauterp", "--quiet"}, cmds...)
	command := exec.Command(s.ExoCLIExecutable, cmds...)
	output, err := command.CombinedOutput()
	s.Assert().NoError(err, string(output))

	return string(output)
}
