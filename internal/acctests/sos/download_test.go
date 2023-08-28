package sos_test

import "os"

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
