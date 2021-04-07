package utils

import (
	"fmt"
	"strings"
)

func TrimLong(tag string, maxLength int) string {
	l := len(tag)
	if l > maxLength {
		tag = fmt.Sprintf("%s..%s", tag[:(maxLength/2)-1], tag[l-(maxLength/2)+1:])
	}
	return tag
}

func MakeSummary(key string, uid string, up bool) string {
	maxLength := 20
	var UID string
	if up {
		UID = strings.ToUpper(uid)
	} else {
		UID = strings.ToLower(uid)
	}
	return fmt.Sprintf("%.*s(%.10s)", maxLength, TrimLong(key, maxLength), UID)
}
