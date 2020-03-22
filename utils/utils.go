package utils

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"
)

var rng = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
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

// RandomString returns a random string of n runes
func RandomString(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A rng.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, rng.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rng.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

// ContainsAll returns true if haystack contains all substrings in needles
func ContainsAll(haystack string, needles []string) bool {
	for _, v := range needles {
		if !strings.Contains(haystack, v) {
			return false
		}
	}
	return true
}
