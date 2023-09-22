package sos_test

import "fmt"

func (s *SOSSuite) TestDownloadSingleObject() {
	s.Execute(SOSTest{
		Steps: []Step{
			{
				PreparedFiles: LocalFiles{
					"file1.txt": "expected content",
				},
				Commands: []string{
					"exo storage upload {prepDir}file1.txt {bucket}",
					"exo storage download -r {bucket} {downloadDir}",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "expected content",
				},
			},
			{
				PreparedFiles: LocalFiles{
					"file1.txt": "this file should not get written",
				},
				Commands: []string{
					"exo storage upload {prepDir}file1.txt {bucket}",
					"exo storage download -r {bucket} {downloadDir}",
				},
				ExpectErrorInCommandNr: 2,
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "expected content",
				},
			},
			{
				PreparedFiles: LocalFiles{
					"file1.txt": "this file should get written",
				},
				Commands: []string{
					"exo storage upload {prepDir}file1.txt {bucket}",
					"exo storage download -f -r {bucket} {downloadDir}",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "this file should get written",
				},
			},
			{
				Description: "check if latest object can be renamed",
				PreparedFiles: LocalFiles{
					"file1.txt": "new new expected content",
				},
				Commands: []string{
					"exo storage upload {prepDir}file1.txt {bucket}",
					"exo storage download {bucket}/file1.txt {downloadDir}/file1-new.txt",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt":     "this file should get written",
					"file1-new.txt": "new new expected content",
				},
			},
		},
	})
}

func (s *SOSSuite) TestDownloadSingleVersionedObject() {
	s.Execute(SOSTest{
		Steps: []Step{
			{
				PreparedFiles: LocalFiles{
					"file1.txt": "original content",
				},
				Commands: []string{
					"exo storage bucket versioning enable {bucket}",
					"exo storage upload {prepDir}file1.txt {bucket}",
				},
				ExpectedDownloadFiles: LocalFiles{},
			},
			{
				Description: "check if latest object is downloaded",
				PreparedFiles: LocalFiles{
					"file1.txt": "new expected content",
				},
				Commands: []string{
					"exo storage upload {prepDir}file1.txt {bucket}",
					"exo storage download -f -r {bucket} {downloadDir}",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "new expected content",
				},
			},
			{
				Description:   "check if v0 can be downloaded",
				PreparedFiles: LocalFiles{},
				Commands: []string{
					"exo storage download -f --only-versions v0 {bucket}/file1.txt {downloadDir}",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "original content",
				},
			},
			{
				Description:   "check if v1 can be explicitly downloaded",
				PreparedFiles: LocalFiles{},
				Commands: []string{
					"exo storage download -f --only-versions v1 {bucket}/file1.txt {downloadDir}",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "new expected content",
				},
			},
			{
				Description:                    "check if v0 can be downloaded and renamed",
				PreparedFiles:                  LocalFiles{},
				ClearDownloadDirBeforeCommands: true,
				Commands: []string{
					"exo storage download -f --only-versions v0 {bucket}/file1.txt {downloadDir}/file1-v0.txt",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1-v0.txt": "original content",
				},
			},
		},
	})
}

func (s *SOSSuite) TestDownloadMultipleObjects() {
	fmt.Println(s.T().Name())
	s.Execute(SOSTest{
		Steps: []Step{
			{
				Description: "check that multiple files can be downloaded",
				PreparedFiles: LocalFiles{
					"file1.txt": "expected content 1",
					"file2.txt": "expected content 2",
				},
				Commands: []string{
					"exo storage upload {prepDir}file1.txt {bucket}",
					"exo storage upload {prepDir}file2.txt {bucket}",
					"exo storage download -r {bucket} {downloadDir}",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "expected content 1",
					"file2.txt": "expected content 2",
				},
			},
			{
				Description: "check that multiple files can be uploaded",
				PreparedFiles: LocalFiles{
					"file1.txt": "expected content 1",
					"file2.txt": "expected content 2",
				},
				Commands: []string{
					"exo storage upload -r {prepDir} {bucket}",
					"exo storage download -r {bucket} {downloadDir}",
				},
				ClearDownloadDirBeforeCommands: true,
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "expected content 1",
					"file2.txt": "expected content 2",
				},
			},
			{
				Description: "check that a directory can be uploaded and downloaded",
				PreparedFiles: LocalFiles{
					"dir/file1.txt": "expected content 1",
					"dir/file2.txt": "expected content 2",
				},
				Commands: []string{
					"exo storage upload -r {prepDir} {bucket}",
					"exo storage download -r {bucket} {downloadDir}",
				},
				ClearDownloadDirBeforeCommands: true,
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt":     "expected content 1",
					"file2.txt":     "expected content 2",
					"dir/file1.txt": "expected content 1",
					"dir/file2.txt": "expected content 2",
				},
			},
			{
				Description:   "check for error if directory download doesn't end in slash",
				PreparedFiles: LocalFiles{},
				Commands: []string{
					"exo storage download -r {bucket} {downloadDir}newDir",
				},
				ClearDownloadDirBeforeCommands: true,
				ExpectErrorInCommandNr:         1,
				ExpectedDownloadFiles:          LocalFiles{},
			},
			{
				Description:   "check that a directory can be downloaded and renamed",
				PreparedFiles: LocalFiles{},
				Commands: []string{
					"exo storage upload -r {prepDir} {bucket}",
					"exo storage download -r {bucket} {downloadDir}newDir/",
				},
				ClearDownloadDirBeforeCommands: true,
				ExpectedDownloadFiles: LocalFiles{
					"newDir/file1.txt":     "expected content 1",
					"newDir/file2.txt":     "expected content 2",
					"newDir/dir/file1.txt": "expected content 1",
					"newDir/dir/file2.txt": "expected content 2",
				},
			},
		},
	})
}
