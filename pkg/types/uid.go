package types

import (
	"encoding/hex"
	"github.com/google/uuid"
)

const (
	ShortUIDLength = 10
)

func newUniqueID() string {
	u, err := uuid.NewRandom()
	if err != nil {
		panic(err) // panic because it cannot happen
	}
	b, err := u.MarshalBinary()
	if err != nil {
		panic(err) // panic because it cannot happen
	}
	return hex.EncodeToString(b)
}

func ShortenUID(uid string) string {
	return uid[:ShortUIDLength]
}

// NewDUID creates a new DUID.
func NewDUID() string {
	return newUniqueID()
}

// ValidateUID validate UID from string.
func ValidateUID(uidStr string) bool {
	uid, err := uuid.Parse(uidStr)
	if err != nil {
		return false
	}
	_, err = uid.MarshalBinary()
	if err != nil {
		return false
	}
	return true
}

// NewCUID creates a new CUID
func NewCUID() string {
	return newUniqueID()
}

// NewNilCUID creates an instance of Nil CUID.
func NewNilCUID() string {
	var bin = make([]byte, 16)
	return hex.EncodeToString(bin)
}
