package types

import (
	"bytes"
	"github.com/google/uuid"
)

// UniqueID is unique ID in the format of UUID.
type UniqueID []byte

func newUniqueID() UniqueID {
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

func (its UniqueID) String() string {
	uid, err := uuid.FromBytes(its)
	if err != nil {
		return "fail to make string to uuid"
	}
	return uid.String()
}

// CompareUID compares two UIDs.
func CompareUID(a, b UniqueID) int {
	return bytes.Compare(a, b)
}
