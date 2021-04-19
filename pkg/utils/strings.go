package utils

import (
	"fmt"
)

const maxLength = 20

func MakeShort(tag string, maxLength int) string {
	l := len(tag)
	if l > maxLength {
		tag = fmt.Sprintf("%s..%s", tag[:(maxLength/2)-1], tag[l-(maxLength/2)+1:])
	}
	return fmt.Sprintf("%.*s", maxLength, tag)
}

func MakeDefaultShort(tag string) string {
	return MakeShort(tag, maxLength)
}
