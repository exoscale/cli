package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

const (
	binaryFilePrefix    = "file://"
	binaryRawFilePrefix = "fileb://"
)

// BinaryValue is a binary CLI parameter. It holds decoded bytes and accepts
// the following input forms:
//   - "<base64>": a base64-encoded value
//   - "file://<path>": a file containing a base64-encoded value
//   - "fileb://<path>": a file containing the raw binary value
type BinaryValue []byte

var _ pflag.Value = (*BinaryValue)(nil)

// Set parses v according to the supported input forms.
func (b *BinaryValue) Set(v string) error {
	switch {
	case strings.HasPrefix(v, binaryRawFilePrefix):
		data, err := os.ReadFile(strings.TrimPrefix(v, binaryRawFilePrefix))
		if err != nil {
			return fmt.Errorf("unable to read file: %w", err)
		}
		*b = data
		return nil

	case strings.HasPrefix(v, binaryFilePrefix):
		data, err := os.ReadFile(strings.TrimPrefix(v, binaryFilePrefix))
		if err != nil {
			return fmt.Errorf("unable to read file: %w", err)
		}
		v = strings.TrimSpace(string(data))
	}

	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return fmt.Errorf("invalid base64 value: %w", err)
	}
	*b = data

	return nil
}

// String returns the base64 encoding of the value.
func (b *BinaryValue) String() string {
	if b == nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(*b)
}

// Type returns the type name displayed in the flags usage.
func (b *BinaryValue) Type() string { return "binary" }
