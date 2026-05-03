package dsl

import "strings"

func splitFirst(s, sep string) [2]string {
	idx := strings.IndexByte(s, sep[0])
	return [2]string{s[:idx], s[idx+len(sep):]}
}
