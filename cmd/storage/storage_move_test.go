package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateStorageTransferArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		recursive bool
		err       string
	}{
		{
			name: "single object",
			args: []string{"sos://bucket/source", "sos://bucket/destination"},
		},
		{
			name: "prefix to another bucket root",
			args: []string{"sos://bucket/source/", "sos://other-bucket/"},
		},
		{
			name: "identical source and destination",
			args: []string{"sos://bucket/source", "sos://bucket/source"},
			err:  "source and destination must differ",
		},
		{
			name: "destination within source prefix",
			args: []string{"sos://bucket/source/", "sos://bucket/source/backup/"},
			err:  "source and destination prefixes must not overlap",
		},
		{
			name: "destination is parent of source prefix",
			args: []string{"sos://bucket/source/nested/", "sos://bucket/source/"},
			err:  "source and destination prefixes must not overlap",
		},
		{
			name:      "recursive destination within source prefix",
			args:      []string{"sos://bucket/source", "sos://bucket/source-backup"},
			recursive: true,
			err:       "source and destination prefixes must not overlap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStorageTransferArgs(tt.args, tt.recursive)
			if tt.err == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.err)
		})
	}
}

func TestStorageTransferAliases(t *testing.T) {
	assert.Contains(t, storageCopyCmd.Aliases, "cp")
	assert.Contains(t, storageMoveCmd.Aliases, "mv")
}
