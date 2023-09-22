package sos_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type LocalFiles map[string]string

type Step struct {
	Description                    string
	PreparedFiles                  LocalFiles
	ClearDownloadDirBeforeCommands bool
	Commands                       []string

	// 0 means don't expect an error, counting starts at 1
	ExpectErrorInCommandNr int
	ExpectedDownloadFiles  LocalFiles
}

type SOSTest struct {
	Steps []Step
}

func emptyDirectory(dirPath string) error {
	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Read all directory entries
	entries, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	// Loop through the entries and remove them
	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())

		// Remove file or directory
		if entry.IsDir() {
			err = os.RemoveAll(entryPath)
		} else {
			err = os.Remove(entryPath)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func registerFile(s *SOSSuite, files LocalFiles, prefix string) fs.WalkDirFunc {
	return func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		s.NoError(err)

		localPath := strings.TrimPrefix(path, prefix)
		files[localPath] = string(content)
		return nil
	}
}

func (s *SOSSuite) Execute(test SOSTest) {
	for stepNr, step := range test.Steps {
		s.T().Logf("step number: %d %q", stepNr, step.Description)
		for filename, content := range step.PreparedFiles {
			s.writeFile(filename, content)
		}

		if step.ClearDownloadDirBeforeCommands && !s.NoError(emptyDirectory(s.DownloadDir)) {
			return
		}

		for i, command := range step.Commands {
			commandNr := i + 1

			_, err := s.exo(command)

			errorExpected := step.ExpectErrorInCommandNr > 0 && step.ExpectErrorInCommandNr == commandNr
			gotError := err != nil

			switch {
			case errorExpected && !gotError:
				s.Fail("expected error in command nr ", step.ExpectErrorInCommandNr)

				return
			case !errorExpected && gotError:
				s.NoError(err)

				return
			case errorExpected == gotError:
			}
		}

		downloadDir := LocalFiles{}
		err := filepath.WalkDir(s.DownloadDir, registerFile(s, downloadDir, s.DownloadDir))
		s.NoError(err)
		nFiles := len(downloadDir)
		fmt.Printf("downloadDir: %v\n", downloadDir)

		actualFileNumberMismatches := !s.Equal(len(step.ExpectedDownloadFiles), nFiles, "number of actual files doesn't match number of expected files")
		if actualFileNumberMismatches {
			return
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
