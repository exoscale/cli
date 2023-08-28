package sos_test

import "os"

type LocalFiles map[string]string

type Step struct {
	PreparedFiles         LocalFiles
	Commands              []string
	ExpectedDownloadFiles LocalFiles
}

type SOSTest struct {
	Steps []Step
}

func (s *SOSSuite) Execute(test SOSTest) {
	for _, step := range test.Steps {
		for filename, content := range step.PreparedFiles {
			s.writeFile(filename, content)
		}

		for _, command := range step.Commands {
			s.exo(command)
		}

		files, err := os.ReadDir(s.DownloadDir)
		if !s.NoError(err) {
			return
		}

		if !s.Equal(len(step.ExpectedDownloadFiles), len(files), "number of actual files doesn't match number of expected files") {
			return
		}

		downloadDir := LocalFiles{}
		for _, file := range files {
			content, err := os.ReadFile(s.DownloadDir + file.Name())
			if !s.NoError(err) {
				return
			}

			downloadDir[file.Name()] = string(content)
		}

		for expectedFilename, expectedContent := range step.ExpectedDownloadFiles {
			actualContent, ok := downloadDir[expectedFilename]
			if !s.True(ok) {
				return
			}

			if !s.Equal(expectedContent, actualContent) {
				return
			}
		}
	}
}
