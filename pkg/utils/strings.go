package utils

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
	"strings"
)

func TrimLong(tag string, maxLength int) string {
	l := len(tag)
	if l > maxLength {
		tag = fmt.Sprintf("%s..%s", tag[:(maxLength/2)-1], tag[l-(maxLength/2)+1:])
	}
	return tag
}

func MakeSummary(key string, uid []byte, up bool) string {
	maxLength := 20
	UID := types.UIDtoString(uid)
	if up {
		UID = strings.ToUpper(UID)
	} else {
		UID = strings.ToLower(UID)
	}
	return fmt.Sprintf("%.*s(%.10s)", maxLength, TrimLong(key, maxLength), UID)
}
