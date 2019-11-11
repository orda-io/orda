package model

import (
	"bytes"
	"github.com/knowhunger/ortoo/commons/log"
)

//DUID is the unique ID of datatype
type DUID UniqueID

//NewDUID creates a new DUID
func NewDUID() (DUID, error) {
	u, err := newUniqueID()
	if err != nil {
		return nil, log.OrtooErrorf(err, "fail to generate datatype UID")
	}
	return DUID(u), nil
}

func (d DUID) String() string {
	return UniqueID(d).String()
}

func (d DUID) Compare(o []byte) int {
	return bytes.Compare(UniqueID(d), o)
}
