package userdata

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	maxUserDataLength = 32768
)

func GetUserDataFromFile(path string, compress bool) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	userData, err := EncodeUserData(data, compress)
	if err != nil {
		return "", err
	}

	if len(userData) >= maxUserDataLength {
		return "", fmt.Errorf("user-data maximum allowed length is %d bytes", maxUserDataLength)
	}

	return userData, nil
}

func EncodeUserData(data []byte, compress bool) (string, error) {
	if compress {
		b := new(bytes.Buffer)
		gz := gzip.NewWriter(b)

		if _, err := gz.Write(data); err != nil {
			return "", err
		}
		if err := gz.Flush(); err != nil {
			return "", err
		}
		if err := gz.Close(); err != nil {
			return "", err
		}

		data = b.Bytes()
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func DecodeUserData(data string) (string, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	gz, err := gzip.NewReader(bytes.NewReader(base64Decoded))
	if err != nil {
		// User data are not compressed, returning as-is.
		if errors.Is(err, gzip.ErrHeader) {
			return string(base64Decoded), nil
		}

		return "", err
	}
	defer gz.Close()

	userData, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}

	return string(userData), nil
}
