package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

const maxLength = 20

// MakeShort returns a short tag
func MakeShort(tag string, maxLength int) string {
	l := len(tag)
	if l > maxLength {
		tag = fmt.Sprintf("%s..%s", tag[:(maxLength/2)-1], tag[l-(maxLength/2)+1:])
	}
	return fmt.Sprintf("%.*s", maxLength, tag)
}

// MakeDefaultShort makes short
func MakeDefaultShort(tag string) string {
	return MakeShort(tag, maxLength)
}

// HashSum returns the hash sum of arguments
func HashSum(args ...interface{}) string {
	hash := md5.New()
	for _, arg := range args {
		if stringer, ok := arg.(fmt.Stringer); ok {
			hash.Write([]byte(stringer.String()))
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
