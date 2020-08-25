package types

import (
	"bytes"
	"encoding/hex"
	"github.com/google/uuid"
)

const (
	shortStringLength = 10
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
	return ToUID(its)
}

func (its UID) ShortString() string {
	return ToShortUID(its)
}

// CompareUID compares two UIDs.
func CompareUID(a, b UID) int {
	return bytes.Compare(a, b)
}

func ToShortUID(uid []byte) string {
	return hex.EncodeToString(uid)[:shortStringLength]
}

func ToUID(uid []byte) string {
	return hex.EncodeToString(uid)
}
