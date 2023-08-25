package main

import (
	"context"
	"fmt"
	"io"
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
}

func (s *SOSSuite) SetupTest() {
	ctx := context.Background()

	// TODO build the cli
	// if err := exec.Command("go", "build").Run(); err != nil {
	// 	fmt.Println("Error building CLI:", err)
	// 	return
	// }

	// Create test and download directories
	if err := os.MkdirAll(s.PrepDir, 0755); err != nil {
		fmt.Println("Error creating test directory:", err)
		return
	}
	if err := os.MkdirAll(s.DownloadDir, 0755); err != nil {
		fmt.Println("Error creating download directory:", err)
		return
	}

	var caCerts io.Reader

	cfg, err := config.LoadDefaultConfig(
		ctx,
		append([]func(*config.LoadOptions) error{},
			config.WithRegion(zone),

			config.WithEndpointResolver(aws.EndpointResolverFunc(
				func(service, region string) (aws.Endpoint, error) {
					sosURL := strings.Replace("https://sos-{zone}.exo.io", "{zone}", zone, 1)
					return aws.Endpoint{
						URL:           sosURL,
						SigningRegion: zone,
					}, nil
				})),

			config.WithCustomCABundle(caCerts),
		)...)
	s.Assert().NoError(err)

	// TODO assert that bucket doesn't exist
	//
	// TODO create bucket
	// input := &s3.CreateBucketInput{
	// 	Bucket: &bucketName,
	// }

	// TODO enable versioning

	_ = s3.NewFromConfig(cfg)
	//svc := s3.NewFromConfig(cfg)
	// _, err = svc.CreateBucket(ctx, input)
	s.Assert().NoError(err)
}

func (s *SOSSuite) TearDownTest() {
	if err := os.RemoveAll(s.PrepDir); err != nil {
		fmt.Println("Error cleaning up test directory:", err)
	}

	if err := os.RemoveAll(s.DownloadDir); err != nil {
		fmt.Println("Error cleaning up download directory:", err)
	}
}

func TestSOSSuite(t *testing.T) {
	integDir := "integdir/"
	s := &SOSSuite{
		BucketName:       "integ-bucket",
		ExoCLIExecutable: "../../../cli",

		PrepDir:     integDir + "prep/",
		DownloadDir: integDir + "downloads/",
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
}

func (s *SOSSuite) exo(args string) string {
	cmds := strings.Split(args, " ")

	cmds = append([]string{"-A", "exosauterp", "--quiet"}, cmds...)
	command := exec.Command(s.ExoCLIExecutable, cmds...)
	output, err := command.CombinedOutput()
	s.Assert().NoError(err, string(output))

	return string(output)
}
