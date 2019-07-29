package model

import (
	"github.com/google/uuid"
	"github.com/knowhunger/ortoo/commons/log"
)

type uniqueID []byte

func newUniqueID() (uniqueID, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, log.OrtooError(err, "fail to generate unique ID")
	}
	b, err := u.MarshalBinary()
	if err != nil {
		return nil, log.OrtooError(err, "fail to generate unique ID")
	}
	return uniqueID(b), nil
}

func (u uniqueID) String() string {
	uid, err := uuid.FromBytes([]byte(u))
	if err != nil {
		return "fail to make string to uuid"
	}
	return uid.String()
}
