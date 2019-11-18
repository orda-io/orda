package model

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/knowhunger/ortoo/commons/log"
)

type UniqueID []byte

func newUniqueID() (UniqueID, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate unique ID")
	}
	b, err := u.MarshalBinary()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate unique ID")
	}
	return b, nil
}

func (u UniqueID) String() string {
	uid, err := uuid.FromBytes(u)
	if err != nil {
		return "fail to make string to uuid"
	}
	return uid.String()
}

func CompareUID(a, b UniqueID) int {
	return bytes.Compare(a, b)
}
