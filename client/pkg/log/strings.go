package log

import "fmt"

const maxLength = 20

// MakeShort returns a short tag
func MakeShort(tag string, maxLength int) string {
	l := len(tag)
	if l > maxLength {
		tag = fmt.Sprintf("%s♻%s", tag[:(maxLength/2)-1], tag[l-(maxLength/2)+1:])
	}
	return fmt.Sprintf("%.*s", maxLength, tag)
}

// MakeDefaultShort makes short
func MakeDefaultShort(tag string) string {
	return MakeShort(tag, maxLength)
}
