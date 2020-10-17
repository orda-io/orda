package types

import (
	"bytes"
	"encoding/hex"
	"github.com/google/uuid"
)

const (
	ShortUIDLength = 10
)

// UID is unique ID in the format of UUID.
type UID []byte

func newUniqueID() UID {
	u, err := uuid.NewUUID()
	if err != nil {
		panic(err) // panic because it cannot happen
	}
	b, err := u.MarshalBinary()
	if err != nil {
		panic(err) // panic because it cannot happen
	}
	return b
}

func (its UID) String() string {
	return UIDtoString(its)
}

// ShortString returns a short string.
func (its UID) ShortString() string {
	return UIDtoShortString(its)
}

// CompareUID compares two UIDs.
func CompareUID(a, b UID) int {
	return bytes.Compare(a, b)
}

// UIDtoShortString returns a short UID string.
func UIDtoShortString(uid []byte) string {
	return hex.EncodeToString(uid)[:ShortUIDLength]
}

// UIDtoString returns a string of UID.
func UIDtoString(uid []byte) string {
	return hex.EncodeToString(uid)
}

func ShortenUIDString(uid string) string {
	return uid[:ShortUIDLength]
}
