package common

import (
	"crypto/md5"
	"encoding/hex"
)

// TruncateString returns s truncated n characters
func TruncateString(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// GetMD5Hash returns the md5 hash of a string, as a string
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
