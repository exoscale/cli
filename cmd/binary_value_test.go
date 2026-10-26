package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BinaryValue_Set(t *testing.T) {
	raw := []byte{0x00, 0xff, 0x10, 'h', 'i', '\n'}
	b64 := "AP8QaGkK"

	dir := t.TempDir()
	binFile := filepath.Join(dir, "data.bin")
	require.NoError(t, os.WriteFile(binFile, raw, 0o600))
	b64File := filepath.Join(dir, "data.b64")
	require.NoError(t, os.WriteFile(b64File, []byte(b64+"\n"), 0o600))
	badB64File := filepath.Join(dir, "bad.b64")
	require.NoError(t, os.WriteFile(badB64File, []byte("not base64!"), 0o600))

	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr string
	}{
		{name: "base64 value", input: b64, want: raw},
		{name: "empty value", input: "", want: []byte{}},
		{name: "invalid base64 value", input: "not base64!", wantErr: "invalid base64 value"},
		{name: "fileb", input: "fileb://" + binFile, want: raw},
		{name: "fileb missing file", input: "fileb://" + filepath.Join(dir, "missing"), wantErr: "unable to read file"},
		{name: "file", input: "file://" + b64File, want: raw},
		{name: "file invalid base64", input: "file://" + badB64File, wantErr: "invalid base64 value"},
		{name: "file missing file", input: "file://" + filepath.Join(dir, "missing"), wantErr: "unable to read file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b BinaryValue
			err := b.Set(tt.input)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, []byte(b))
		})
	}
}

func Test_BinaryValue_String(t *testing.T) {
	b := BinaryValue{0x00, 0xff, 0x10, 'h', 'i', '\n'}
	assert.Equal(t, "AP8QaGkK", b.String())

	var nilValue *BinaryValue
	assert.Equal(t, "", nilValue.String())
}

func Test_BinaryValue_Type(t *testing.T) {
	var b BinaryValue
	assert.Equal(t, "binary", b.Type())
}
