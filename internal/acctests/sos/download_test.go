package sos_test

import (
	"os"
)

type FileDirectory map[string]string

type Step struct {
	PreparedFiles         FileDirectory
	Commands              []string
	ExpectedDownloadFiles FileDirectory
}

type SOSTest struct {
	Steps []Step
}

func (s *SOSSuite) TestDownloadFiles() {
	tests := SOSTest{
		Steps: []Step{
			{
				PreparedFiles: FileDirectory{
					"file1.txt": "expected content",
				},
				Commands: []string{
					"storage upload {prepDir}file1.txt {bucket}",
					"storage download -r {bucket} {downloadDir}",
				},
				ExpectedDownloadFiles: FileDirectory{
					"file1.txt": "expected content",
				},
			},
		},
	}

	// TODO abort on the first error
	for _, step := range tests.Steps {
		for filename, content := range step.PreparedFiles {
			s.writeFile(filename, content)
		}

		for _, command := range step.Commands {
			s.exoCMD(command)
		}

		files, err := os.ReadDir(s.DownloadDir)
		s.Assert().NoError(err)
		s.Equal(len(step.ExpectedDownloadFiles), len(files), "number of actual files doesn't match number of expected files")

		downState := FileDirectory{}
		for _, file := range files {
			content, err := os.ReadFile(s.DownloadDir + file.Name())
			s.Assert().NoError(err)

			downState[file.Name()] = string(content)
		}

		for expectedFilename, expectedContent := range step.ExpectedDownloadFiles {
			actualContent, ok := downState[expectedFilename]
			s.True(ok)

			s.Equal(expectedContent, actualContent)
		}
	}

	// file1Name := "file1.txt"
	// originalContent := "original content"

	// s.writeFile(file1Name, originalContent)
	// s.uploadFile(file1Name)

	// expectedContent := "expected new content"
	// s.writeFile(file1Name, expectedContent)
	// s.uploadFile(file1Name)

	// s.downloadVersion(s.findLatestVersion())

	// files, err := os.ReadDir(s.DownloadDir)
	// s.Assert().NoError(err)
	// s.Equal(1, len(files))
	// s.Equal(file1Name, files[0].Name())

	// file1Content, err := os.ReadFile(s.DownloadDir + file1Name)
	// s.Assert().NoError(err)
	// s.Equal(expectedContent, string(file1Content))
}
