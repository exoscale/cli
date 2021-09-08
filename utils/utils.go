package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// RandStringBytes Generate random string of n bytes
func RandStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
