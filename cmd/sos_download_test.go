package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDestinationExists(t *testing.T) {
	t.Run("Should return true if the localFilePath exists", func(t *testing.T) {
		tempfile, err := os.CreateTemp("", "temp.txt")
		if err != nil {
			t.Errorf("failed to create tempfile: %v", err)
		}
		defer os.Remove(tempfile.Name())

		exists, err := destinationExists(tempfile.Name(), "")
		if err != nil {
			t.Errorf("destinationExists returned error: %v", err)
		}
		if !exists {
			t.Errorf("destinationExists should return true for path: %s", tempfile.Name())
		}
	})

	t.Run("Should return true if the localFilePath already contains the object", func(t *testing.T) {
		tempdir, err := os.MkdirTemp("", "subdir")
		if err != nil {
			t.Errorf("failed to create tempdir: %v", err)
		}
		defer os.RemoveAll(tempdir)

		tempfile, err := os.CreateTemp(tempdir, "temp.txt")
		if err != nil {
			t.Errorf("failed to create tempfile: %v", err)
		}
		// Cleanup happens when we remove the parent dir

		objectName := filepath.Base(tempfile.Name())
		exists, err := destinationExists(tempdir, objectName)
		if err != nil {
			t.Errorf("destinationExists returned error: %v", err)
		}
		if !exists {
			t.Errorf("destinationExists should return true for path: %s and object: %s", tempdir, objectName)
		}
	})

	t.Run("Should return false if localFilePath does not exist", func(t *testing.T) {
		tempfile, _ := os.CreateTemp("", "temp.txt")
		os.Remove(tempfile.Name())

		exists, err := destinationExists(tempfile.Name(), "")
		if err != nil {
			t.Errorf("destinationExists returned an error: %v", err)
		}
		if exists {
			t.Errorf("destinationExists should return false for path: %s", tempfile.Name())
		}
	})

	t.Run("Should return false if localFilePath is a folder without the object", func(t *testing.T) {
		tempdir, err := os.MkdirTemp("", "subdir")
		if err != nil {
			t.Errorf("failed to create tempdir: %v", err)
		}
		defer os.RemoveAll(tempdir)

		exists, err := destinationExists(tempdir, "test.txt")
		if err != nil {
			t.Errorf("destinationExists returned an error: %v", err)
		}
		if exists {
			t.Errorf("destinationExists should return false for path: %s", tempdir)
		}
	})
}
