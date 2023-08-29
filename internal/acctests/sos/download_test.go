package sos_test

func (s *SOSSuite) TestDownloadFiles() {
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
		},
	})
}

func (s *SOSSuite) TestDownloadOverwrittenVersionedFile() {
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
		},
	})
}
