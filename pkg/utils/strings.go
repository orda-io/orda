package utils

import (
	"fmt"
	"github.com/knowhunger/ortoo/pkg/types"
)

func TrimLong(tag string, maxLength int) string {
	l := len(tag)
	if l > maxLength {
		tag = fmt.Sprintf("%s..%s", tag[:(maxLength/2)-1], tag[l-(maxLength/2)+1:])
	}
	return tag
}

func MakeSummary(key string, uid []byte, prefix string) string {
	maxLength := 20
	return fmt.Sprintf("%s:%.*s(%.10s)", prefix, maxLength, TrimLong(key, maxLength), types.UIDtoString(uid))
}
