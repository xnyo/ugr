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

// FindString takes a slice and looks for an element in it. If found it will
// return its index, otherwise it will return -1 and false.
func FindString(haystack []string, needle string) (int, bool) {
	for i, v := range haystack {
		if v == needle {
			return i, true
		}
	}
	return -1, false
}
