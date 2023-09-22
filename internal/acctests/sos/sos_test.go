package sos_test

import (
	"os"
	"path/filepath"
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

		files, err := os.ReadDir(s.DownloadDir)
		if !s.NoError(err) {
			return
		}

		actualFileNumberMismatches := !s.Equal(len(step.ExpectedDownloadFiles), len(files), "number of actual files doesn't match number of expected files")

		downloadDir := LocalFiles{}
		for _, file := range files {
			if actualFileNumberMismatches {
				s.T().Logf("actual file: %s", file)
			}

			content, err := os.ReadFile(s.DownloadDir + file.Name())
			if !s.NoError(err) {
				return
			}

			downloadDir[file.Name()] = string(content)
		}

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
