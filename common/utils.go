package common

// TruncateString returns s truncated n characters
func TruncateString(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
