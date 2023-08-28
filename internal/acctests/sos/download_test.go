package sos_test

func (s *SOSSuite) TestDownloadFiles() {
	test := SOSTest{
		Steps: []Step{
			{
				PreparedFiles: LocalFiles{
					"file1.txt": "expected content",
				},
				Commands: []string{
					"storage upload {prepDir}file1.txt {bucket}",
					"storage download -r {bucket} {downloadDir}",
				},
				ExpectedDownloadFiles: LocalFiles{
					"file1.txt": "expected content",
				},
			},
		},
	}

	s.Execute(test)
}
