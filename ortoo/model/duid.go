package model

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/knowhunger/ortoo/ortoo/log"
)

// DUID is the unique ID of a datatype.
type DUID UniqueID

// NewDUID creates a new DUID.
func NewDUID() (DUID, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate datatype UID")
	}
	return DUID(u), nil
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

func (d DUID) String() string {
	return UniqueID(d).String()
}

// Compare compares a DUID with another.
func (d DUID) Compare(o []byte) int {
	return bytes.Compare(UniqueID(d), o)
}
