package types

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/knowhunger/ortoo/ortoo/log"
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
		return nil, log.OrtooError(err)
	}
	b, err := uid.MarshalBinary()
	if err != nil {
		return nil, log.OrtooError(err)
	}
	return b, nil
}

func (its DUID) String() string {
	return ToUID(its)
}

func (its DUID) ShortString() string {
	return ToShortUID(its)
}

// Compare compares a DUID with another.
func (its *DUID) Compare(o []byte) int {
	return bytes.Compare(*its, o)
}
