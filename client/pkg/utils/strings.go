package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

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
