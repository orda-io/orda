package types

import (
	"bytes"
	"github.com/google/uuid"
)

// DUID is the unique ID of a datatype.
type DUID UID

// NewDUID creates a new DUID.
func NewDUID() DUID {
	return DUID(newUniqueID())
}

// DUIDFromString creates DUID from string.
func DUIDFromString(duidString string) (DUID, error) {
	uid, err := uuid.Parse(duidString)
	if err != nil {
		return nil, err
	}
	b, err := uid.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (its DUID) String() string {
	return UIDtoString(its)
}

// ShortString returns a short string.
func (its DUID) ShortString() string {
	return UIDtoShortString(its)
}

// Compare compares a DUID with another.
func (its *DUID) Compare(o []byte) int {
	return bytes.Compare(*its, o)
}
