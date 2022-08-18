package types

import (
	"bytes"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	defaultUIDLength    = 16
	defaultIDCharacters = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func newUniqueID() string {
	return gonanoid.Must(defaultUIDLength)
}

// NewUID creates a new DUID.
func NewUID() string {
	return newUniqueID()
}

// ValidateUID validate UID from string.
func ValidateUID(uidStr string) bool {
	if len(uidStr) != defaultUIDLength {
		return false
	}
	for _, c := range uidStr {
		if !strings.Contains(defaultIDCharacters, string(c)) {
			return false
		}
	}
	return true
}

// NewNilUID creates an instance of Nil UID.
func NewNilUID() string {
	var b bytes.Buffer
	for i := 0; i < defaultUIDLength; i++ {
		b.WriteString("0")
	}
	return b.String()
}
